/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"go.kubeguard.dev/guard/util/kubeconfig"

	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
	goauth "golang.org/x/oauth2/google"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

var gauthConfig oauth2.Config

func IssueToken() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
	}()
	klog.Infoln("Oauth2 callback receiver listening on", listener.Addr())

	// https://developers.google.com/identity/protocols/OpenIDConnect#validatinganidtoken
	gauthConfig = oauth2.Config{
		Endpoint:     goauth.Endpoint,
		ClientID:     GoogleOauth2ClientID,
		ClientSecret: GoogleOauth2ClientSecret,
		Scopes:       []string{"openid", "profile", "email"},
		RedirectURL:  "http://" + listener.Addr().String(),
	}
	// PromptSelectAccount allows a user who has multiple accounts at the authorization server
	// to select amongst the multiple accounts that they may have current sessions for.
	// eg: https://developers.google.com/identity/protocols/OpenIDConnect
	promptSelectAccount := oauth2.SetAuthURLParam("prompt", "select_account")
	codeURL := gauthConfig.AuthCodeURL("/", promptSelectAccount)

	klog.Infoln("Auhtorization code URL:", codeURL)

	err = open.Start(codeURL)
	if err != nil {
		return err
	}

	http.HandleFunc("/", handleGoogleAuth)
	return http.Serve(listener, nil)
}

// https://developers.google.com/identity/protocols/OAuth2InstalledApp
func handleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return
	}
	token, err := gauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-content-type-options", "nosniff")
	w.WriteHeader(http.StatusOK)

	err = addAuthInfo(token.Extra("id_token").(string), token.RefreshToken)
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	} else {
		_, _ = w.Write([]byte("Configuration has been written to " + kubeconfig.Path()))
		return
	}
}

func addAuthInfo(idToken, refreshToken string) error {
	authInfo := &clientcmdapi.AuthInfo{
		AuthProvider: &clientcmdapi.AuthProviderConfig{
			Name: "oidc",
			Config: map[string]string{
				"client-id":      GoogleOauth2ClientID,
				"client-secret":  GoogleOauth2ClientSecret,
				"id-token":       idToken,
				"idp-issuer-url": "https://accounts.google.com",
				"refresh-token":  refreshToken,
			},
		},
	}
	email, err := getEmailFromIdToken(idToken)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve email from idToken")
	}
	return kubeconfig.AddAuthInfo(email, authInfo)
}

func getEmailFromIdToken(idToken string) (string, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("oidc: malformed jwt, expected 3 parts got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("oidc: malformed jwt payload: %v", err)
	}

	c := struct {
		Email string `json:"email"`
	}{}

	err = json.Unmarshal(payload, &c)
	if err != nil {
		return "", err
	}
	return c.Email, nil
}

package cmds

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/appscode/go/log"
	"github.com/appscode/go/term"
	"github.com/appscode/guard/lib"
	"github.com/howeyc/gopass"
	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	goauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

func NewCmdGetToken() *cobra.Command {
	var org string
	cmd := &cobra.Command{
		Use:               "token",
		Short:             fmt.Sprintf("Get tokens for %v", lib.SupportedOrgPrintForm()),
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			org = strings.ToLower(org)
			switch org {
			case "github":
				codeURurl := "https://github.com/settings/tokens/new"
				log.Infoln("Github url for personal access tokens:", codeURurl)
				open.Start(codeURurl)
				return
			case "gitlab":
				codeURurl := "https://gitlab.com/profile/personal_access_tokens"
				log.Infoln("Gitlab url for personal access tokens:", codeURurl)
				open.Start(codeURurl)
				return
			case "google":
				getGoogleToken()
				return
			case "appscode":
				getAppscodeToken()
			case "":
				log.Fatalln("Missing organization name. Set flag -o Google|Github.")
			default:
				log.Fatalf("Unknown organization %s.", org)
			}
		},
	}

	cmd.Flags().StringVarP(&org, "organization", "o", org, fmt.Sprintf("Name of Organization (%v).", lib.SupportedOrgPrintForm()))
	return cmd
}

var gauthConfig goauth2.Config

func getGoogleToken() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()
	log.Infoln("Oauth2 callback receiver listening on", listener.Addr())

	// https://developers.google.com/identity/protocols/OpenIDConnect#validatinganidtoken
	gauthConfig = goauth2.Config{
		Endpoint:     google.Endpoint,
		ClientID:     lib.GoogleOauth2ClientID,
		ClientSecret: lib.GoogleOauth2ClientSecret,
		Scopes:       []string{"openid", "profile", "email"},
		RedirectURL:  "http://" + listener.Addr().String(),
	}
	// PromptSelectAccount allows a user who has multiple accounts at the authorization server
	// to select amongst the multiple accounts that they may have current sessions for.
	// eg: https://developers.google.com/identity/protocols/OpenIDConnect
	promptSelectAccount := oauth2.SetAuthURLParam("prompt", "select_account")
	codeURL := gauthConfig.AuthCodeURL("/", promptSelectAccount)

	log.Infoln("Auhtorization code URL:", codeURL)
	open.Start(codeURL)

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

	err = addUserInKubeConfig(token.Extra("id_token").(string), token.RefreshToken)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	} else {
		w.Write([]byte("Configuration has been written to " + KubeConfigPath()))
		return
	}
}

func getAppscodeToken() error {
	teamId := term.Read("Team Id:")
	endpoint := fmt.Sprintf("https://%v.appscode.io", teamId)
	err := open.Start(strings.Join([]string{endpoint, "conduit", "login"}, "/"))

	term.Print("Paste the token here: ")
	tokenBytes, err := gopass.GetPasswdMasked()
	if err != nil {
		term.Fatalln("Failed to retrieve token", err)
	}

	token := string(tokenBytes)
	client := &lib.ConduitClient{
		Url:   strings.Join([]string{endpoint, "api", "user.whoami"}, "/"),
		Token: token,
	}
	result := &lib.WhoAmIResponse{}
	err = client.Call().Into(result)
	if err != nil {
		term.Fatalln("Failed to validate token", err)
	}
	if result.ErrorCode != nil {
		term.Fatalln("Failed to validate token")
	}
	term.Successln("Token successfully generated")

	return err
}

func addUserInKubeConfig(idToken, refreshToken string) error {
	var konfig *clientcmdapi.Config
	if _, err := os.Stat(KubeConfigPath()); err == nil {
		// ~/.kube/config exists
		konfig, err = clientcmd.LoadFromFile(KubeConfigPath())
		if err != nil {
			return err
		}
	} else {
		konfig = &clientcmdapi.Config{
			APIVersion: "v1",
			Kind:       "Config",
			Preferences: clientcmdapi.Preferences{
				Colors: true,
			},
		}
	}

	authInfo := &clientcmdapi.AuthInfo{
		AuthProvider: &clientcmdapi.AuthProviderConfig{
			Name: "oidc",
			Config: map[string]string{
				"client-id":      lib.GoogleOauth2ClientID,
				"client-secret":  lib.GoogleOauth2ClientSecret,
				"id-token":       idToken,
				"idp-issuer-url": "https://accounts.google.com",
				"refresh-token":  refreshToken,
			},
		},
	}
	email, err := getEmailFromIdToken(idToken)
	if err != nil {
		return errors.Wrap(err, "failed to retrive emial from idToken")
	}
	konfig.AuthInfos["google-"+email] = authInfo

	err = os.MkdirAll(filepath.Dir(KubeConfigPath()), 0755)
	if err != nil {
		return err
	}
	err = clientcmd.WriteToFile(*konfig, KubeConfigPath())
	if err != nil {
		return err
	}
	term.Successln("Configuration has been written to", KubeConfigPath())
	return nil
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

func KubeConfigPath() string {
	return homedir.HomeDir() + "/.kube/config"
}

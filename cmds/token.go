package cmds

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	term "github.com/appscode/go/term"
	"github.com/appscode/guard/lib"
	"github.com/howeyc/gopass"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	goauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

// https://developers.google.com/identity/protocols/OAuth2InstalledApp
const (
	googleOauth2ClientID     = "37154062056-220683ek37naab43v23vc5qg01k1j14g.apps.googleusercontent.com"
	googleOauth2ClientSecret = "pB9ITCuMPLj-bkObrTqKbt57"
)

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
		ClientID:     googleOauth2ClientID,
		ClientSecret: googleOauth2ClientSecret,
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"client_id":     googleOauth2ClientID,
		"client_secret": googleOauth2ClientSecret,
		"id_token":      token.Extra("id_token"),
		"refresh_token": token.RefreshToken,
	})
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

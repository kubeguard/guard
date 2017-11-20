package cmds

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	term "github.com/appscode/go/term"
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
		Short:             "Get tokens for Github or Google",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			org = strings.ToLower(org)
			switch org {
			case "github":
				codeURurl := "https://github.com/settings/tokens/new"
				log.Infoln("Github url for personal access tokens:", codeURurl)
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

	cmd.Flags().StringVarP(&org, "organization", "o", org, "Name of Organization (Github or Google).")
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

	gauthConfig = goauth2.Config{
		Endpoint:     google.Endpoint,
		ClientID:     googleOauth2ClientID,
		ClientSecret: googleOauth2ClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/admin.directory.group.readonly"},
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
	json.NewEncoder(w).Encode(token)
}

func getAppscodeToken() error {
	teamId := term.Read("Team Id:")
	endpoint := fmt.Sprintf("https://%v.appscode.io", teamId)
	err := open.Start(strings.Join([]string{endpoint, "conduit", "login"}, "/"))
	return err
}

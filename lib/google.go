package lib

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/oauth2/google"
	gdir "google.golang.org/api/admin/directory/v1"
	gauth "google.golang.org/api/oauth2/v1"
	auth "k8s.io/api/authentication/v1"
)

type GoogleOptions struct {
	ServiceAccountJsonFile string
	AdminEmail             string
}

const (
	googleIssuerUrl1 = "https://accounts.google.com"
	googleIssuerUrl2 = "accounts.google.com"

	// https://developers.google.com/identity/protocols/OAuth2InstalledApp
	GoogleOauth2ClientID     = "37154062056-220683ek37naab43v23vc5qg01k1j14g.apps.googleusercontent.com"
	GoogleOauth2ClientSecret = "pB9ITCuMPLj-bkObrTqKbt57"
)

func (s *GoogleOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.ServiceAccountJsonFile, "google.sa-json-file", s.ServiceAccountJsonFile, "Path to Google service account json file")
	fs.StringVar(&s.AdminEmail, "google.admin-email", s.AdminEmail, "Email of G Suite administrator")
}

func (s GoogleOptions) ToArgs() []string {
	var args []string

	if s.ServiceAccountJsonFile != "" {
		args = append(args, fmt.Sprintf("--google.sa-json-file=/etc/guard/auth/sa.json"))
	}
	if s.AdminEmail != "" {
		args = append(args, fmt.Sprintf("--google.admin-email=%s", s.AdminEmail))
	}

	return args
}

func (s Server) checkGoogle(name, token string) (auth.TokenReview, int) {
	authSvc, err := gauth.New(http.DefaultClient)
	if err != nil {
		return Error(fmt.Sprintf("Failed to create oauth2/v1 api client for domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
	}
	// ref: https://github.com/google/google-api-go-client/issues/160#issuecomment-345155421
	r1, err := authSvc.Tokeninfo().IdToken(token).Do()
	if err != nil {
		return Error(fmt.Sprintf("Failed to load user info for domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
	}
	// verify step 2-5
	// https://developers.google.com/identity/protocols/OpenIDConnect#validatinganidtoken
	if r1.Issuer != googleIssuerUrl1 && r1.Issuer != googleIssuerUrl2 {
		return Error(fmt.Sprintf("Issuer url didn't match. Expected %v or %v, got %v", googleIssuerUrl1, googleIssuerUrl2, r1.Issuer)), http.StatusUnauthorized
	}

	if r1.ExpiresIn <= 0 {
		return Error(fmt.Sprintf("token is expired")), http.StatusUnauthorized
	}

	if r1.Audience != GoogleOauth2ClientID {
		return Error(fmt.Sprintf("Expected client ID %v, got %v", GoogleOauth2ClientID, r1.Audience)), http.StatusUnauthorized
	}

	if !strings.HasSuffix(r1.Email, "@"+name) {
		return Error(fmt.Sprintf("User is not a member of domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
	}
	resp := auth.TokenReview{}
	resp.Status = auth.TokenReviewStatus{
		User: auth.UserInfo{
			Username: r1.Email,
			UID:      r1.UserId,
		},
	}

	if s.Google.ServiceAccountJsonFile != "" {
		sa, err := ioutil.ReadFile(s.Google.ServiceAccountJsonFile)
		if err != nil {
			return Error(fmt.Sprintf("Failed to load service account json file %s. Reason: %v.", s.Google.ServiceAccountJsonFile, err)), http.StatusUnauthorized
		}

		cfg, err := google.JWTConfigFromJSON(sa, gdir.AdminDirectoryGroupReadonlyScope)
		if err != nil {
			return Error(fmt.Sprintf("Failed to create JWT config from service account json file %s. Reason: %v.", s.Google.ServiceAccountJsonFile, err)), http.StatusUnauthorized
		}

		// https://admin.google.com/ManageOauthClients
		// ref: https://developers.google.com/admin-sdk/directory/v1/guides/delegation
		// Note: Only users with access to the Admin APIs can access the Admin SDK Directory API, therefore your service account needs to impersonate one of those users to access the Admin SDK Directory API.
		cfg.Subject = s.Google.AdminEmail
		client := cfg.Client(context.Background())

		svc, err := gdir.New(client)
		if err != nil {
			return Error(fmt.Sprintf("Failed to create admin/directory/v1 client for domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
		}

		var groups []string
		var pageToken string
		for {
			r2, err := svc.Groups.List().UserKey(r1.Email).Domain(name).PageToken(pageToken).Do()
			if err != nil {
				return Error(fmt.Sprintf("Failed to load user's groups for domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
			}
			for _, group := range r2.Groups {
				groups = append(groups, group.Email)
			}
			if r2.NextPageToken == "" {
				break
			}
			pageToken = r2.NextPageToken
		}
		resp.Status.User.Groups = groups
	}

	resp.Status.Authenticated = true
	return resp, http.StatusOK
}

package lib

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2/google"
	gdir "google.golang.org/api/admin/directory/v1"
	gauth "google.golang.org/api/oauth2/v1"
	auth "k8s.io/api/authentication/v1"
)

const (
	googleIssuerUrl = "https://accounts.google.com"

	// https://developers.google.com/identity/protocols/OAuth2InstalledApp
	GoogleOauth2ClientID     = "37154062056-220683ek37naab43v23vc5qg01k1j14g.apps.googleusercontent.com"
	GoogleOauth2ClientSecret = "pB9ITCuMPLj-bkObrTqKbt57"
)

type GoogleOptions struct {
	ServiceAccountJsonFile string
	AdminEmail             string
}

type GoogleClient struct {
	GoogleOptions
	verifier *oidc.IDTokenVerifier
	ctx      context.Context
	service  *gdir.Service
}

type ExtendedTokenInfo struct {
	gauth.Tokeninfo
	HD string `json:"hd"`
}

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

func NewGoogleClient(opts GoogleOptions, domain string) (*GoogleClient, error) {
	g := &GoogleClient{
		GoogleOptions: opts,
		ctx:           context.Background(),
	}

	var err error
	provider, err := oidc.NewProvider(g.ctx, googleIssuerUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to create oidc provider for google. Reason: %v.", err)
	}

	g.verifier = provider.Verifier(&oidc.Config{
		ClientID: GoogleOauth2ClientID,
	})

	if opts.ServiceAccountJsonFile != "" {
		sa, err := ioutil.ReadFile(opts.ServiceAccountJsonFile)
		if err != nil {
			return nil, fmt.Errorf("Failed to load service account json file %s. Reason: %v.", opts.ServiceAccountJsonFile, err)
		}

		cfg, err := google.JWTConfigFromJSON(sa, gdir.AdminDirectoryGroupReadonlyScope)
		if err != nil {
			return nil, fmt.Errorf("Failed to create JWT config from service account json file %s. Reason: %v.", opts.ServiceAccountJsonFile, err)
		}

		// https://admin.google.com/ManageOauthClients
		// ref: https://developers.google.com/admin-sdk/directory/v1/guides/delegation
		// Note: Only users with access to the Admin APIs can access the Admin SDK Directory API, therefore your service account needs to impersonate one of those users to access the Admin SDK Directory API.
		cfg.Subject = opts.AdminEmail
		client := cfg.Client(context.Background())

		g.service, err = gdir.New(client)
		if err != nil {
			return nil, fmt.Errorf("Failed to create admin/directory/v1 client for domain %s. Reason: %v.", domain, err)
		}
	}
	return g, nil
}

func (g *GoogleClient) checkGoogle(name, token string) (auth.TokenReview, int) {
	idToken, err := g.verifier.Verify(g.ctx, token)
	if err != nil {
		return Error(fmt.Sprintf("Failed to verify token for google. Reason: %v.", err)), http.StatusUnauthorized
	}

	user := ExtendedTokenInfo{}

	err = idToken.Claims(&user)
	if err != nil {
		return Error(fmt.Sprintf("Failed to get claim from token. Reason: %v", err)), http.StatusUnauthorized
	}

	// https://developers.google.com/identity/protocols/OpenIDConnect#validatinganidtoken
	if user.HD != name {
		return Error(fmt.Sprintf("User is not a member of domain %s.", name)), http.StatusUnauthorized
	}

	resp := auth.TokenReview{}
	resp.Status = auth.TokenReviewStatus{
		User: auth.UserInfo{
			Username: user.Email,
			UID:      user.UserId,
		},
	}

	if g.ServiceAccountJsonFile != "" {
		var groups []string
		var pageToken string

		for {
			r2, err := g.service.Groups.List().UserKey(user.Email).Domain(name).PageToken(pageToken).Do()
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

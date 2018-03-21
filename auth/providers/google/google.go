package google

import (
	"context"
	"io/ioutil"

	"github.com/appscode/guard/auth"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	gdir "google.golang.org/api/admin/directory/v1"
	gauth "google.golang.org/api/oauth2/v1"
	authv1 "k8s.io/api/authentication/v1"
)

const (
	OrgType = "google"

	googleIssuerUrl = "https://accounts.google.com"
	// https://developers.google.com/identity/protocols/OAuth2InstalledApp
	GoogleOauth2ClientID     = "37154062056-220683ek37naab43v23vc5qg01k1j14g.apps.googleusercontent.com"
	GoogleOauth2ClientSecret = "pB9ITCuMPLj-bkObrTqKbt57"
)

func init() {
	auth.SupportedOrgs = append(auth.SupportedOrgs, OrgType)
}

type Authenticator struct {
	Options
	verifier *oidc.IDTokenVerifier
	ctx      context.Context
	service  *gdir.Service
}

type TokenInfo struct {
	gauth.Tokeninfo
	HD string `json:"hd"`
}

func New(opts Options, domain string) (*Authenticator, error) {
	g := &Authenticator{
		Options: opts,
		ctx:     context.Background(),
	}

	provider, err := oidc.NewProvider(g.ctx, googleIssuerUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create oidc provider for google")
	}

	g.verifier = provider.Verifier(&oidc.Config{
		ClientID: GoogleOauth2ClientID,
	})

	if opts.ServiceAccountJsonFile != "" {
		sa, err := ioutil.ReadFile(opts.ServiceAccountJsonFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load service account json file %s", opts.ServiceAccountJsonFile)
		}

		cfg, err := google.JWTConfigFromJSON(sa, gdir.AdminDirectoryGroupReadonlyScope)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create JWT config from service account json file %s", opts.ServiceAccountJsonFile)
		}

		// https://admin.google.com/ManageOauthClients
		// ref: https://developers.google.com/admin-sdk/directory/v1/guides/delegation
		// Note: Only users with access to the Admin APIs can access the Admin SDK Directory API, therefore your service account needs to impersonate one of those users to access the Admin SDK Directory API.
		cfg.Subject = opts.AdminEmail
		client := cfg.Client(context.Background())

		g.service, err = gdir.New(client)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create admin/directory/v1 client for domain %s", domain)
		}
	}
	return g, nil
}

// https://developers.google.com/identity/protocols/OpenIDConnect#validatinganidtoken
func (g *Authenticator) Check(name, token string) (*authv1.UserInfo, error) {
	idToken, err := g.verifier.Verify(g.ctx, token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify token for google")
	}

	info := TokenInfo{}

	err = idToken.Claims(&info)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get claim from token")
	}

	if info.HD != name {
		return nil, errors.Errorf("user is not a member of domain %s", name)
	}

	resp := &authv1.UserInfo{
		Username: info.Email,
		UID:      info.UserId,
	}

	if g.ServiceAccountJsonFile != "" {
		var groups []string
		var pageToken string

		for {
			r2, err := g.service.Groups.List().UserKey(info.Email).Domain(name).PageToken(pageToken).Do()
			if err != nil {
				return nil, errors.Wrapf(err, "failed to load user's groups for domain %s", name)
			}
			for _, group := range r2.Groups {
				groups = append(groups, group.Email)
			}
			if r2.NextPageToken == "" {
				break
			}
			pageToken = r2.NextPageToken
		}
		resp.Groups = groups
	}

	return resp, nil
}

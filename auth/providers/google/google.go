package google

import (
	"context"

	"github.com/appscode/guard/auth"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
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
	verifier   *oidc.IDTokenVerifier
	ctx        context.Context
	service    *gdir.Service
	token      string
	domainName string
}

type TokenInfo struct {
	gauth.Tokeninfo
	HD string `json:"hd"`
}

func New(opts Options, domain, token string) (auth.Interface, error) {
	g := &Authenticator{
		Options:    opts,
		ctx:        context.Background(),
		domainName: domain,
		token:      token,
	}

	provider, err := oidc.NewProvider(g.ctx, googleIssuerUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create oidc provider for google")
	}

	g.verifier = provider.Verifier(&oidc.Config{
		ClientID: GoogleOauth2ClientID,
	})

	if opts.ServiceAccountJsonFile != "" {
		client := opts.jwtConfig.Client(context.Background())

		g.service, err = gdir.New(client)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create admin/directory/v1 client for domain %s", domain)
		}
	}
	return g, nil
}

func (g Authenticator) UID() string {
	return OrgType
}

// https://developers.google.com/identity/protocols/OpenIDConnect#validatinganidtoken
func (g *Authenticator) Check() (*authv1.UserInfo, error) {
	idToken, err := g.verifier.Verify(g.ctx, g.token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify token for google")
	}

	info := TokenInfo{}

	err = idToken.Claims(&info)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get claim from token")
	}

	if info.HD != g.domainName {
		return nil, errors.Errorf("user is not a member of domain %s", g.domainName)
	}

	resp := &authv1.UserInfo{
		Username: info.Email,
		UID:      info.UserId,
	}

	if g.ServiceAccountJsonFile != "" {
		var groups []string
		var pageToken string

		for {
			r2, err := g.service.Groups.List().UserKey(info.Email).Domain(g.domainName).PageToken(pageToken).Do()
			if err != nil {
				return nil, errors.Wrapf(err, "failed to load user's groups for domain %s", g.domainName)
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

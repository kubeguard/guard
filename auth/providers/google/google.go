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

	"github.com/appscode/guard/auth"

	oidc "github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	gdir "google.golang.org/api/admin/directory/v1"
	gauth "google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
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
	domainName string
}

type TokenInfo struct {
	gauth.Tokeninfo
	HD string `json:"hd"`
}

func New(opts Options, domain string) (auth.Interface, error) {
	g := &Authenticator{
		Options:    opts,
		ctx:        context.Background(),
		domainName: domain,
	}

	provider, err := oidc.NewProvider(g.ctx, googleIssuerUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create oidc provider for google")
	}

	g.verifier = provider.Verifier(&oidc.Config{
		ClientID: GoogleOauth2ClientID,
	})

	if opts.ServiceAccountJsonFile != "" {
		ctx := context.Background()
		g.service, err = gdir.NewService(ctx, option.WithTokenSource(opts.jwtConfig.TokenSource(ctx)))
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
func (g *Authenticator) Check(token string) (*authv1.UserInfo, error) {
	idToken, err := g.verifier.Verify(g.ctx, token)
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

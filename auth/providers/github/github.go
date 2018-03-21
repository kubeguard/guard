package github

import (
	"context"
	"fmt"

	"github.com/appscode/guard/auth"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	authv1 "k8s.io/api/authentication/v1"
)

const (
	OrgType = "github"
)

func init() {
	auth.SupportedOrgs = append(auth.SupportedOrgs, OrgType)
}

type Authenticator struct {
	Client  *github.Client
	ctx     context.Context
	OrgName string // Github organization name
}

func New(name string) auth.Interface {
	g := &Authenticator{
		ctx:     context.Background(),
		OrgName: name,
	}
	g.Client = github.NewClient(oauth2.NewClient(g.ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)))

	return g
}

func (g Authenticator) UID() string {
	return OrgType
}

func (g *Authenticator) Check(token string) (*authv1.UserInfo, error) {
	mem, _, err := g.Client.Organizations.GetOrgMembership(g.ctx, "", g.OrgName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to check user's membership in Org %s", g.OrgName)
	}

	resp := &authv1.UserInfo{
		Username: mem.User.GetLogin(),
		UID:      fmt.Sprintf("%d", mem.User.GetID()),
	}

	var groups []string
	page := 1
	pageSize := 25
	for {
		teams, _, err := g.Client.Organizations.ListUserTeams(g.ctx, &github.ListOptions{Page: page, PerPage: pageSize})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load user's teams for Org %s", g.OrgName)
		}
		for _, team := range teams {
			if team.Organization.GetLogin() == g.OrgName {
				groups = append(groups, team.GetName())
			}
		}
		if len(teams) < pageSize {
			break
		}
		page++
	}
	resp.Groups = groups
	return resp, nil
}

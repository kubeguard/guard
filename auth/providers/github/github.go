package github

import (
	"context"
	"fmt"

	"github.com/appscode/guard/auth"
	"github.com/google/go-github/v25/github"
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
	opts    Options
	ctx     context.Context
	OrgName string // Github organization name
}

func New(opts Options, name string) auth.Interface {
	g := &Authenticator{
		opts:    opts,
		ctx:     context.Background(),
		OrgName: name,
	}

	return g
}

func (g Authenticator) UID() string {
	return OrgType
}

func (g *Authenticator) Check(token string) (*authv1.UserInfo, error) {
	var (
		client *github.Client
		err    error
	)

	if g.opts.BaseUrl != "" {
		client, err = github.NewEnterpriseClient(g.opts.BaseUrl, "", oauth2.NewClient(g.ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Github enterprise client")
		}
	} else {
		client = github.NewClient(oauth2.NewClient(g.ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)))
	}

	mem, _, err := client.Organizations.GetOrgMembership(g.ctx, "", g.OrgName)
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
		teams, _, err := client.Teams.ListUserTeams(g.ctx, &github.ListOptions{Page: page, PerPage: pageSize})
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

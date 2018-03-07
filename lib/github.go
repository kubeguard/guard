package lib

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	auth "k8s.io/api/authentication/v1"
)

type GithubClient struct {
	Client  *github.Client
	Ctx     context.Context
	OrgName string // Github organization name
}

func NewGithubClient(name, token string) *GithubClient {
	g := &GithubClient{
		Ctx:     context.Background(),
		OrgName: name,
	}
	g.Client = github.NewClient(oauth2.NewClient(g.Ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)))

	return g
}

func (g *GithubClient) checkGithub() (auth.TokenReview, int) {
	mem, _, err := g.Client.Organizations.GetOrgMembership(g.Ctx, "", g.OrgName)
	if err != nil {
		return Error(fmt.Sprintf("Failed to check user's membership in Org %s. Reason: %v.", g.OrgName, err)), http.StatusUnauthorized
	}
	resp := auth.TokenReview{}
	resp.Status = auth.TokenReviewStatus{
		User: auth.UserInfo{
			Username: mem.User.GetLogin(),
			UID:      fmt.Sprintf("%d", mem.User.GetID()),
		},
	}

	var groups []string
	page := 1
	pageSize := 25
	for {
		teams, _, err := g.Client.Organizations.ListUserTeams(g.Ctx, &github.ListOptions{Page: page, PerPage: pageSize})
		if err != nil {
			return Error(fmt.Sprintf("Failed to load user's teams for Org %s. Reason: %v.", g.OrgName, err)), http.StatusUnauthorized
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
	resp.Status.User.Groups = groups
	resp.Status.Authenticated = true
	return resp, http.StatusOK
}

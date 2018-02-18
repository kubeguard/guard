package lib

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	auth "k8s.io/api/authentication/v1"
)

func checkGithub(name, token string) (auth.TokenReview, int) {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)))

	mem, _, err := client.Organizations.GetOrgMembership(ctx, "", name)
	if err != nil {
		return Error(fmt.Sprintf("Failed to check user's membership in Org %s. Reason: %v.", name, err)), http.StatusUnauthorized
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
		teams, _, err := client.Organizations.ListUserTeams(ctx, &github.ListOptions{Page: page, PerPage: pageSize})
		if err != nil {
			return Error(fmt.Sprintf("Failed to load user's teams for Org %s. Reason: %v.", name, err)), http.StatusUnauthorized
		}
		for _, team := range teams {
			if team.Organization.GetLogin() == name {
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

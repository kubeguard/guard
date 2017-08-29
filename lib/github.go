package lib

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/github"
	"github.com/tamalsaha/go-oneliners"
	"golang.org/x/oauth2"
	auth "k8s.io/client-go/pkg/apis/authentication/v1beta1"
)

func checkGithub(name, token string) (auth.TokenReview, int) {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)))

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return Error(fmt.Sprintf("Failed to load user's Github profile. Reason: %v.", err)), http.StatusUnauthorized
	}
	data := auth.TokenReview{}
	data.Status = auth.TokenReviewStatus{
		User: auth.UserInfo{
			Username: user.GetLogin(),
			UID:      strconv.Itoa(user.GetID()),
		},
	}

	groups := []string{}
	page := 1
	pageSize := 25
	for {
		teams, _, err := client.Organizations.ListUserTeams(ctx, &github.ListOptions{Page: page, PerPage: pageSize})
		if err != nil {
			return Error(fmt.Sprintf("Failed to load user's teams. Reason: %v.", err)), http.StatusUnauthorized
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
	data.Status.User.Groups = groups
	return data, http.StatusOK
}

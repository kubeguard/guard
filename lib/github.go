package lib

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	auth "k8s.io/client-go/pkg/apis/authentication/v1beta1"
)

func checkGithub(w http.ResponseWriter, name, token string) {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)))

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		Error(w, fmt.Sprintf("Failed to load user's Github profile. Reason: %v.", err), http.StatusUnauthorized)
		return
	}
	data := NewTokenReview()
	data.Status = auth.TokenReviewStatus{
		User: auth.UserInfo{
			Username: user.GetLogin(),
			UID:      strconv.Itoa(user.GetID()),
		},
	}

	groups := []string{}
	page := 1
	for {
		teams, _, err := client.Organizations.ListUserTeams(ctx, &github.ListOptions{Page: page})
		if err != nil {
			Error(w, fmt.Sprintf("Failed to load user's teams. Reason: %v.", err), http.StatusUnauthorized)
			return
		}
		for _, team := range teams {
			if team.Organization.GetLogin() == name {
				groups = append(groups, team.GetName())
			}
		}
		page++
	}
	data.Status.User.Groups = groups
	Write(w, data)
}

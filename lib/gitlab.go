package lib

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/xanzy/go-gitlab"
	auth "k8s.io/api/authentication/v1"
)

type GitlabClient struct {
	Client *gitlab.Client
}

func NewGitlabClient(token string) *GitlabClient {
	return &GitlabClient{
		Client: gitlab.NewClient(nil, token),
	}
}

func (g *GitlabClient) checkGitLab() (auth.TokenReview, int) {
	user, _, err := g.Client.Users.CurrentUser()
	if err != nil {
		return Error(err.Error()), http.StatusUnauthorized
	}

	resp := auth.TokenReview{}
	resp.Status = auth.TokenReviewStatus{
		User: auth.UserInfo{
			Username: user.Username,
			UID:      strconv.Itoa(user.ID),
		},
	}

	var groups []string
	// https://docs.gitlab.com/ee/api/README.html#pagination
	page := 1
	pageSize := 20
	for {
		list, _, err := g.Client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions: gitlab.ListOptions{Page: page, PerPage: pageSize},
		})
		if err != nil {
			return Error(fmt.Sprintf("Failed to load groups. Reason: %v", err)), http.StatusBadRequest
		}
		for _, g := range list {
			groups = append(groups, g.Name)
		}
		if len(list) < pageSize {
			break
		}
		page++
	}

	resp.Status.User.Groups = groups
	resp.Status.Authenticated = true
	return resp, http.StatusOK
}

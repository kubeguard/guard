package lib

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/xanzy/go-gitlab"
	auth "k8s.io/api/authentication/v1beta1"
)

func checkGitLab(token string) (auth.TokenReview, int) {
	client := gitlab.NewClient(nil, token)

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return Error(err.Error()), http.StatusUnauthorized
	}

	groupList, _, err := client.Groups.ListGroups(nil)
	if err != nil {
		return Error(fmt.Sprintf("failed to check informaiton of Group. Reason: %v", err)), http.StatusBadRequest
	}
	groups := make([]string, 0, len(groupList))
	for _, g := range groupList {
		groups = append(groups, g.Name)
	}

	resp := auth.TokenReview{}
	resp.Status = auth.TokenReviewStatus{
		User: auth.UserInfo{
			Username: user.Username,
			UID:      strconv.Itoa(user.ID),
			Groups:   groups,
		},
	}

	resp.Status.Authenticated = true
	return resp, http.StatusOK
}

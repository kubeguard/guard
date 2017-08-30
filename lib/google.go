package lib

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	"golang.org/x/oauth2"
	gdir "google.golang.org/api/admin/directory/v1"
	gauth "google.golang.org/api/oauth2/v1"
	auth "k8s.io/client-go/pkg/apis/authentication/v1beta1"
)

func checkGoogle(name, token string) (auth.TokenReview, int) {
	ctx := context.Background()
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	authSvc, err := gauth.New(client)
	if err != nil {
		return Error(fmt.Sprintf("Failed to create oauth2/v1 api client for domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
	}
	r1, err := authSvc.Userinfo.Get().Do()
	if err != nil {
		msg := fmt.Sprintf("Failed to load user info for domain %s. Reason: %v.", name, err)
		log.Errorln(msg)
		return Error(msg), http.StatusUnauthorized
	}

	data := auth.TokenReview{}
	data.Status = auth.TokenReviewStatus{
		User: auth.UserInfo{
			Username: r1.Email,
			UID:      r1.Id,
		},
	}

	svc, err := gdir.New(client)
	if err != nil {
		return Error(fmt.Sprintf("Failed to create admin/directory/v1 client for domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
	}

	groups := []string{}
	var pageToken string
	for {
		r2, err := svc.Groups.List().UserKey(r1.Email).PageToken(pageToken).Do()
		if err != nil {
			return Error(fmt.Sprintf("Failed to load user's groups for domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
		}
		for _, group := range r2.Groups {
			if strings.HasSuffix(group.Email, "@"+name) {
				groups = append(groups, group.Email)
			}
		}
		if r2.NextPageToken == "" {
			break
		}
		pageToken = r2.NextPageToken
	}
	data.Status.User.Groups = groups
	data.Status.Authenticated = true
	return data, http.StatusOK
}

package lib

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	gdir "google.golang.org/api/admin/directory/v1"
	gauth "google.golang.org/api/oauth2/v1"
	auth "k8s.io/client-go/pkg/apis/authentication/v1beta1"
)

func checkGoogle(w http.ResponseWriter, name, token string) {
	ctx := context.Background()
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	authSvc, err := gauth.New(client)
	if err != nil {
		Error(w, fmt.Sprintf("Failed to create oauth2/v1 api client. Reason: %v.", err), http.StatusUnauthorized)
		return
	}
	r1, err := authSvc.Userinfo.Get().Do()
	if err != nil {
		Error(w, fmt.Sprintf("Failed to load user info. Reason: %v.", err), http.StatusUnauthorized)
		return
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
		Error(w, fmt.Sprintf("Failed to create admin/directory/v1 client. Reason: %v.", err), http.StatusUnauthorized)
		return
	}

	groups := []string{}
	var pageToken string
	for {
		r2, err := svc.Groups.List().UserKey(r1.Email).PageToken(pageToken).Do()
		if err != nil {
			Error(w, fmt.Sprintf("Failed to load user's groups. Reason: %v.", err), http.StatusUnauthorized)
			return
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
	Write(w, data)
}

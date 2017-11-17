package lib

import (
	"context"
	"fmt"
	"net/http"

	api "github.com/appscode/api/auth/v1beta1"
	"github.com/appscode/api/dtypes"
	appscode "github.com/appscode/client"
	_env "github.com/appscode/go/env"
	auth "k8s.io/api/authentication/v1beta1"
)

func checkAppscode(name, token string) (auth.TokenReview, int) {
	ctx := context.Background()
	options := appscode.NewOption(_env.ProdApiServer)
	options.UserAgent("appscode/guard")
	/*namespace := strings.Split(name, ".")
	if len(namespace) != 3 {
		return Error(fmt.Sprintf("Failed to detect namespace from domain: %v", name)), http.StatusUnauthorized
	}*/
	options = options.BearerAuth(name, token)
	client, err := appscode.New(options)
	if err != nil {
		return Error(fmt.Sprintf("Failed to create oauth2/v1 api client for domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
	}

	user, err := client.Authentication().Conduit().WhoAmI(ctx, &dtypes.VoidRequest{})
	if err != nil {
		return Error(fmt.Sprintf("Failed to load user info for domain %s.appscode.com. Reason: %v.", name, err)), http.StatusUnauthorized
	}

	projects, err := client.Authentication().Project().List(ctx, &api.ProjectListRequest{
		WithMember: false,
		Members:    []string{user.User.Phid},
	})
	if err != nil {
		return Error(fmt.Sprintf("Failed to load user's teams for Org %s. Reason: %v.", name, err)), http.StatusUnauthorized
	}
	data := auth.TokenReview{}
	data.Status = auth.TokenReviewStatus{
		User: auth.UserInfo{
			Username: user.User.UserName,
			UID:      user.User.Phid,
		},
	}

	groups := []string{}
	for _, project := range projects.Projects {
		groups = append(groups, project.Name)
	}
	data.Status.User.Groups = groups
	data.Status.Authenticated = true
	return data, http.StatusOK
}

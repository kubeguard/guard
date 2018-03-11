package gitlab

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	auth "k8s.io/api/authentication/v1"
)

const (
	OrgType = "gitlab"
)

type Authenticator struct {
	Client *gitlab.Client
}

func New(token string) *Authenticator {
	return &Authenticator{
		Client: gitlab.NewClient(nil, token),
	}
}

func (g *Authenticator) Check() (*auth.UserInfo, error) {
	user, _, err := g.Client.Users.CurrentUser()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp := &auth.UserInfo{
		Username: user.Username,
		UID:      strconv.Itoa(user.ID),
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
			return nil, errors.Wrap(err, "failed to load groups")
		}
		for _, g := range list {
			groups = append(groups, g.Name)
		}
		if len(list) < pageSize {
			break
		}
		page++
	}

	resp.Groups = groups
	return resp, nil
}

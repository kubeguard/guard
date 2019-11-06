/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package gitlab

import (
	"strconv"

	"github.com/appscode/guard/auth"

	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
	authv1 "k8s.io/api/authentication/v1"
)

const (
	OrgType = "gitlab"
)

func init() {
	auth.SupportedOrgs = append(auth.SupportedOrgs, OrgType)
}

type Authenticator struct {
	opts Options
}

func New(opts Options) auth.Interface {
	return &Authenticator{
		opts: opts,
	}
}

func (g Authenticator) UID() string {
	return OrgType
}

func (g *Authenticator) Check(token string) (*authv1.UserInfo, error) {
	client := gitlab.NewClient(nil, token)
	if g.opts.BaseUrl != "" {
		client.SetBaseURL(g.opts.BaseUrl)
	}

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp := &authv1.UserInfo{
		Username: user.Username,
		UID:      strconv.Itoa(user.ID),
	}

	var groups []string
	// https://docs.gitlab.com/ee/api/README.html#pagination
	page := 1
	pageSize := 20
	for {
		list, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions: gitlab.ListOptions{Page: page, PerPage: pageSize},
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to load groups")
		}
		for _, entry := range list {
			if g.opts.UseGroupID {
				groups = append(groups, strconv.Itoa(entry.ID))
			} else {
				groups = append(groups, entry.FullPath)
			}
		}
		if len(list) < pageSize {
			break
		}
		page++
	}

	resp.Groups = groups
	return resp, nil
}

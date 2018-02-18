package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	api "appscode.com/api/auth/v1beta1"
	"appscode.com/api/dtypes"
	"appscode.com/client-go"
	_env "github.com/appscode/go/env"
	auth "k8s.io/api/authentication/v1"
)

type WhoAmIResponse struct {
	ErrorCode interface{}      `json:"error_code"`
	ErrorInfo interface{}      `json:"error_info"`
	Result    *ConduitUserData `json:"result"`
}

type ConduitUserData struct {
	Image        string   `json:"image"`
	Phid         string   `json:"phid"`
	PrimaryEmail string   `json:"primaryEmail"`
	RealName     string   `json:"realName"`
	Roles        []string `json:"roles"`
	URI          string   `json:"uri"`
	UserName     string   `json:"userName"`
}

type ConduitClient struct {
	Url  string
	err  error
	body []byte

	Token string
}

func checkAppscode(name, token string) (auth.TokenReview, int) {
	ctx := context.Background()
	options := client.NewOption(_env.ProdApiServer)
	options.UserAgent("c/guard")
	/*namespace := strings.Split(name, ".")
	if len(namespace) != 3 {
		return Error(fmt.Sprintf("Failed to detect namespace from domain: %v", name)), http.StatusUnauthorized
	}*/
	options = options.BearerAuth(name, token)
	c, err := client.New(options)
	if err != nil {
		return Error(fmt.Sprintf("Failed to create oauth2/v1 api c for domain %s. Reason: %v.", name, err)), http.StatusUnauthorized
	}

	user, err := c.Authentication().Conduit().WhoAmI(ctx, &dtypes.VoidRequest{})
	if err != nil {
		return Error(fmt.Sprintf("Failed to load user info for domain %s.c.com. Reason: %v.", name, err)), http.StatusUnauthorized
	}

	projects, err := c.Authentication().Project().List(ctx, &api.ProjectListRequest{
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

	var groups []string
	for _, project := range projects.Projects {
		groups = append(groups, project.Name)
	}
	data.Status.User.Groups = groups
	data.Status.Authenticated = true
	return data, http.StatusOK
}

func (p *ConduitClient) Call() *ConduitClient {
	client := http.Client{}
	form := url.Values{}
	form.Add("api.token", p.Token)

	phReq, err := http.NewRequest("POST", p.Url, strings.NewReader(form.Encode()))
	if err != nil {
		p.err = err
		return p
	}
	phReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	phResp, err := client.Do(phReq)
	if err != nil {
		p.err = err
		return p
	}
	message, err := ioutil.ReadAll(phResp.Body)
	if err != nil {
		p.err = err
		return p
	}
	p.body = message
	return p
}

func (p *ConduitClient) Into(i interface{}) error {
	if p.err != nil {
		return p.err
	}

	err := json.Unmarshal(p.body, i)
	return err
}

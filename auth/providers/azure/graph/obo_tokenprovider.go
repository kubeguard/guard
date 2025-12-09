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

package graph

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

type oboTokenProvider struct {
	name         string
	client       *http.Client
	clientID     string
	clientSecret string
	scope        string
	loginURL     string
}

// NewOBOTokenProvider returns a TokenProvider that implements OAuth On-Behalf-Of flow on Azure Active Directory
// https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-on-behalf-of-flow
func NewOBOTokenProvider(clientID, clientSecret, loginURL, scope string) TokenProvider {
	return &oboTokenProvider{
		name:         "OBOTokenProvider",
		client:       httpclient.DefaultHTTPClient,
		clientID:     clientID,
		clientSecret: clientSecret,
		scope:        scope,
		loginURL:     loginURL,
	}
}

func (u *oboTokenProvider) Name() string { return u.name }

func (u *oboTokenProvider) Acquire(ctx context.Context, token string) (AuthResponse, error) {
	authResp := AuthResponse{}
	form := url.Values{}
	form.Set("client_id", u.clientID)
	form.Set("client_secret", u.clientSecret)
	form.Set("assertion", token)
	form.Set("requested_token_use", "on_behalf_of")
	form.Set("scope", u.scope)
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")

	req, err := http.NewRequest(http.MethodPost, u.loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return authResp, errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if klog.V(10).Enabled() {
		cmd, _ := http2curl.GetCurlCommand(req)
		klog.V(10).Infoln(cmd)
	}

	resp, err := u.client.Do(req.WithContext(ctx))
	if err != nil {
		return authResp, errors.Wrap(err, "failed to send request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return authResp, errors.Errorf("request %s failed with status code: %d and response: %s", req.URL.Path, resp.StatusCode, string(data))
	}
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return authResp, errors.Wrapf(err, "failed to decode response for request %s", req.URL.Path)
	}

	return authResp, nil
}

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

type clientCredentialTokenProvider struct {
	name         string
	client       *http.Client
	clientID     string
	clientSecret string
	scope        string
	loginURL     string
}

// NewClientCredentialTokenProvider returns a TokenProvider that implements OAuth client credential flow on Azure Active Directory
// https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-client-creds-grant-flow#get-a-token
func NewClientCredentialTokenProvider(clientID, clientSecret, loginURL, scope string) TokenProvider {
	return &clientCredentialTokenProvider{
		name:         "ClientCredentialTokenProvider",
		client:       httpclient.DefaultHTTPClient,
		clientID:     clientID,
		clientSecret: clientSecret,
		scope:        scope,
		loginURL:     loginURL,
	}
}

func (u *clientCredentialTokenProvider) Name() string { return u.name }

func (u *clientCredentialTokenProvider) Acquire(ctx context.Context, token string) (AuthResponse, error) {
	authResp := AuthResponse{}
	form := url.Values{}
	form.Set("client_id", u.clientID)
	form.Set("client_secret", u.clientSecret)
	form.Set("scope", u.scope)
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, u.loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return authResp, errors.Wrap(err, "fail to create request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if klog.V(10).Enabled() {
		cmd, _ := http2curl.GetCurlCommand(req)
		klog.V(10).Infoln(cmd)
	}

	resp, err := u.client.Do(req.WithContext(ctx))
	if err != nil {
		return authResp, errors.Wrap(err, "fail to send request")
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

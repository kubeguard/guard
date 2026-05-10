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
	"bytes"
	"context"
	"io"
	"net/http"

	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/pkg/errors"
)

type aksTokenProvider struct {
	name     string
	client   *http.Client
	tokenURL string
	tenantID string
	resource string
}

// NewAKSTokenProvider returns a TokenProvider that implements On-Behalf-Of flow using AKS first party service
func NewAKSTokenProvider(tokenURL, tenantID string) TokenProvider {
	return &aksTokenProvider{
		name:     "AKSTokenProvider",
		client:   httpclient.DefaultHTTPClient,
		tokenURL: tokenURL,
		tenantID: tenantID,
	}
}

// NewAKSTokenProviderWithResource returns a TokenProvider that requests tokens
// with a specific resource/audience from the OBO service. Used by CheckAccess v2
// to obtain PDP-audience tokens instead of the default ARM-audience tokens.
func NewAKSTokenProviderWithResource(tokenURL, tenantID, resource string) TokenProvider {
	return &aksTokenProvider{
		name:     "AKSTokenProvider",
		client:   httpclient.DefaultHTTPClient,
		tokenURL: tokenURL,
		tenantID: tenantID,
		resource: resource,
	}
}

func (u *aksTokenProvider) Name() string { return u.name }

func (u *aksTokenProvider) Acquire(ctx context.Context, token string) (AuthResponse, error) {
	authResp := AuthResponse{}
	tokenReq := struct {
		TenantID    string `json:"tenantID,omitempty"`
		AccessToken string `json:"accessToken,omitempty"`
		Resource    string `json:"resource,omitempty"`
	}{
		TenantID:    u.tenantID,
		AccessToken: token,
		Resource:    u.resource,
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(tokenReq); err != nil {
		return authResp, errors.Wrap(err, "failed to decode token request")
	}
	req, err := http.NewRequest(http.MethodPost, u.tokenURL, buf)
	if err != nil {
		return authResp, errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.client.Do(req.WithContext(ctx))
	if err != nil {
		return authResp, errors.Wrap(err, "failed to send request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return authResp, errors.Errorf("request failed with status code: %d and response: %s", resp.StatusCode, string(data))
	}
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return authResp, errors.Wrapf(err, "failed to decode response")
	}

	return authResp, nil
}

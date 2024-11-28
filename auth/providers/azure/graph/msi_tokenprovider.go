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
	"strconv"

	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/pkg/errors"
)

const (
	MSIEndpointForARC = "http://127.0.0.1:8421/metadata/identity/oauth2/token?api-version=2018-02-01"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type msiTokenProvider struct {
	name        string
	client      *http.Client
	resource    string
	msiEndpoint string
}

// NewMSITokenProvider returns a TokenProvider that implements OAuth msi flow on Azure Active Directory
func NewMSITokenProvider(msiAudience, msiEndpoint string) TokenProvider {
	return &msiTokenProvider{
		name:        "MSITokenProvider",
		client:      httpclient.DefaultHTTPClient,
		resource:    msiAudience,
		msiEndpoint: msiEndpoint,
	}
}

func (u *msiTokenProvider) Name() string { return u.name }

func (u *msiTokenProvider) Acquire(ctx context.Context, token string) (AuthResponse, error) {
	tokenResp := &TokenResponse{}
	authResp := AuthResponse{}
	var msi_endpoint *url.URL
	msi_endpoint, err := url.Parse(u.msiEndpoint)
	if err != nil {
		return authResp, errors.Wrap(err, "Failed to create msi request for getting token.")
	}
	msi_parameters := msi_endpoint.Query()
	msi_parameters.Add("resource", u.resource)
	msi_endpoint.RawQuery = msi_parameters.Encode()
	req, err := http.NewRequest(http.MethodGet, msi_endpoint.String(), nil)
	if err != nil {
		return authResp, errors.Wrap(err, "Failed to create msi request for getting token.")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Metadata", "true")
	client := httpclient.DefaultHTTPClient
	resp, err := client.Do(req)
	if err != nil {
		return authResp, errors.Wrap(err, "Failed to send msi request for getting token.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return authResp, errors.Errorf("MSI Token Request failed with status code: %d and response: %s", resp.StatusCode, string(data))
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return authResp, errors.Wrapf(err, "Failed to decode response")
	}

	authResp.TokenType = tokenResp.TokenType
	// This value is true only at token creation time, if a cached token is used, its not accurate
	authResp.Expires, err = strconv.Atoi(tokenResp.ExpiresIn)
	if err != nil {
		return authResp, errors.Wrapf(err, "Failed to decode expiry date")
	}
	// This is the actual time the token expires in Unix time
	authResp.ExpiresOn, err = strconv.Atoi(tokenResp.ExpiresOn)
	if err != nil {
		return authResp, errors.Wrapf(err, "Failed to decode expires_on field for token")
	}
	authResp.Token = tokenResp.AccessToken

	return authResp, nil
}

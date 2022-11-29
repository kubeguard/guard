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

package operations

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"go.kubeguard.dev/guard/util/httpclient"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ManagedClusters             = "Microsoft.ContainerService/managedClusters"
	Fleets                      = "Microsoft.ContainerService/fleets"
	ConnectedClusters           = "Microsoft.Kubernetes/connectedClusters"
	OperationsEndpointFormatARC = "%s/providers/Microsoft.Kubernetes/operations?api-version=2021-10-01"
	OperationsEndpointFormatAKS = "%s/providers/Microsoft.ContainerService/operations?api-version=2018-10-31"
	MSIEndpointForARC           = "http://127.0.0.1:8421/metadata/identity/oauth2/token?api-version=2018-02-01"
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

type Display struct {
	Provider    string `json:"provider"`
	Resource    string `json:"resource"`
	Operation   string `json:"operation"`
	Description string `json:"description"`
}

type Operation struct {
	Name         string  `json:"name"`
	Display      Display `json:"display"`
	IsDataAction *bool   `json:"isDataAction,omitempty"`
}

type OperationList struct {
	Value    []Operation `json:"value"`
	NextLink string      `json:"nextLink"`
}

type Resource struct {
	Id         string
	Namespaced bool
	Name       string
	Group      string
	Verb       string
}

type AuthorizationEntity struct {
	Id string `json:"Id"`
}

type AuthorizationActionInfo struct {
	AuthorizationEntity
	IsDataAction bool `json:"IsDataAction"`
}

type DataAction struct {
	ActionInfo           AuthorizationActionInfo
	IsNamespacedResource bool
}

func GetAuthHeaderUsingMSIForARC(resource string) (*TokenResponse, error) {
	tokenResp := &TokenResponse{}
	var msi_endpoint *url.URL
	msi_endpoint, err := url.Parse(MSIEndpointForARC)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create request for getting token.")
	}
	msi_parameters := msi_endpoint.Query()
	msi_parameters.Add("resource", resource)
	msi_endpoint.RawQuery = msi_parameters.Encode()
	req, err := http.NewRequest(http.MethodGet, msi_endpoint.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create request for getting token.")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Metadata", "true")
	client := httpclient.DefaultHTTPClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to send request for getting token.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Errorf("Request failed with status code: %d and response: %s", resp.StatusCode, string(data))
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to decode response")
	}

	return tokenResp, nil
}

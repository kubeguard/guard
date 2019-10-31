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
package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/appscode/guard/auth"
	"github.com/appscode/guard/auth/providers/azure/graph"

	"github.com/Azure/go-autorest/autorest/azure"
	oidc "github.com/coreos/go-oidc"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	authv1 "k8s.io/api/authentication/v1"
)

/*
    ref:
	https://www.youtube.com/watch?v=WygwzN9FfMQ
	https://www.youtube.com/watch?v=ujzrq8Fg9Gc
	https://cloudblogs.microsoft.com/enterprisemobility/2014/12/18/azure-active-directory-now-with-group-claims-and-application-roles/
	https://azure.microsoft.com/en-us/resources/samples/active-directory-dotnet-webapp-groupclaims/
	https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-protocols-openid-connect-code
	https://github.com/kubernetes/kubernetes/pull/43987
	https://github.com/cosmincojocar/kubernetes/blob/682d5ec01f37c65117b2496865cc9bf0cd9e0902/staging/src/k8s.io/client-go/plugin/pkg/client/auth/azure/README.md
*/

const (
	OrgType            = "azure"
	azureUsernameClaim = "upn"
	azureObjectIDClaim = "oid"
)

func init() {
	auth.SupportedOrgs = append(auth.SupportedOrgs, OrgType)
}

var (
	// ErrorClaimNotFound indicates the given key was not found in the claims
	ErrClaimNotFound = fmt.Errorf("claim not found")
)

// claims represents a map of claims provided with a JWT
type claims map[string]interface{}

type Authenticator struct {
	Options
	graphClient *graph.UserInfo
	verifier    *oidc.IDTokenVerifier
	ctx         context.Context
}

type authInfo struct {
	AADEndpoint string
	MSGraphHost string
	Issuer      string
}

func New(opts Options) (auth.Interface, error) {
	c := &Authenticator{
		Options: opts,
		ctx:     context.Background(),
	}
	authInfoVal, err := getAuthInfo(c.Environment, c.TenantID, getMetadata)
	if err != nil {
		return nil, err
	}

	glog.V(3).Infof("Using issuer url: %v", authInfoVal.Issuer)

	provider, err := oidc.NewProvider(c.ctx, authInfoVal.Issuer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create provider for azure")
	}

	c.verifier = provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	c.graphClient, err = graph.New(c.ClientID, c.ClientSecret, c.TenantID, c.UseGroupUID, authInfoVal.AADEndpoint, authInfoVal.MSGraphHost)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ms graph client")
	}
	return c, nil
}

type metadataJSON struct {
	Issuer      string `json:"issuer"`
	MsgraphHost string `json:"msgraph_host"`
}

// https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-convert-app-to-be-multi-tenant
func getMetadata(aadEndpoint, tenantID string) (*metadataJSON, error) {
	metadataURL := aadEndpoint + tenantID + "/.well-known/openid-configuration"
	glog.V(5).Infof("Querying metadata URL: %v", metadataURL)

	response, err := http.Get(metadataURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get metadata url failed with status code: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var metadata metadataJSON
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (s Authenticator) UID() string {
	return OrgType
}

func (s Authenticator) Check(token string) (*authv1.UserInfo, error) {
	idToken, err := s.verifier.Verify(s.ctx, token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify token for azure")
	}

	claims, err := getClaims(idToken)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing claims")
	}

	resp, err := claims.getUserInfo(azureUsernameClaim, azureObjectIDClaim)
	if err != nil {
		return nil, err
	}
	resp.Groups, err = s.graphClient.GetGroups(resp.Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get groups")
	}
	return resp, nil
}

// GetClaims returns a Claims object
func getClaims(token *oidc.IDToken) (claims, error) {
	var c = claims{}
	err := token.Claims(&c)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling claims: %s", err)
	}
	return c, nil
}

// ReviewFromClaims creates a new TokenReview object from the claims object
// the claims object
func (c claims) getUserInfo(usernameClaim, userObjectIDClaim string) (*authv1.UserInfo, error) {
	username, err := c.string(usernameClaim)
	if err != nil && err == ErrClaimNotFound {
		username, err = c.string(userObjectIDClaim)
	}
	if err != nil {
		if err == ErrClaimNotFound {
			return nil, errors.Errorf("username: %s and objectID: %s claims not found", usernameClaim, userObjectIDClaim)
		}
		return nil, errors.Wrap(err, "unable to get username claim")
	}

	return &authv1.UserInfo{Username: username}, nil
}

// String gets a string value from claims given a key. Returns error if
// the key does not exist
func (c claims) string(key string) (string, error) {
	v, ok := c[key]
	if !ok {
		return "", ErrClaimNotFound
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("claim %s is not a string, found %v", key, v)
	}
	return s, nil
}

func getAuthInfo(environment, tenantID string, getMetadata func(string, string) (*metadataJSON, error)) (*authInfo, error) {
	var err error
	env := azure.PublicCloud
	if environment != "" {
		env, err = azure.EnvironmentFromName(environment)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse environment for azure")
		}
	}

	metadata, err := getMetadata(env.ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metadata for azure")
	}

	return &authInfo{
		AADEndpoint: env.ActiveDirectoryEndpoint,
		MSGraphHost: metadata.MsgraphHost,
		Issuer:      metadata.Issuer,
	}, nil
}

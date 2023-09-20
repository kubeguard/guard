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
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.kubeguard.dev/guard/auth"
	"go.kubeguard.dev/guard/auth/providers/azure/graph"
	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/coreos/go-oidc"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/klog/v2"
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

// ErrorClaimNotFound indicates the given key was not found in the claims
var ErrClaimNotFound = fmt.Errorf("claim not found")

// claims represents a map of claims provided with a JWT
type claims map[string]interface{}

type Authenticator struct {
	Options
	graphClient      *graph.UserInfo
	verifier         *oidc.IDTokenVerifier
	popTokenVerifier *PoPTokenVerifier
}

type authInfo struct {
	AADEndpoint string
	MSGraphHost string
	Issuer      string
}

// TODO: combine auth info and issuer into one
var (
	cachedAuthInfo *authInfo
	mutex          = &sync.Mutex{}

	cachedOIDCIssuerProviders map[string]*oidc.Provider = make(map[string]*oidc.Provider)
	cachedOIDCIssuerProvidersMutex = &sync.RWMutex{}
)

func getOIDCIssuerProvider(issuerURL string, issuerGetRetryCount int) (*oidc.Provider, error) {
    // fast path: read from cache
	cachedOIDCIssuerProvidersMutex.RLock()
	cached, ok := cachedOIDCIssuerProviders[issuerURL]
	cachedOIDCIssuerProvidersMutex.RUnlock()
	if ok {
		return cached, nil
	}

	// slow path: construct from remote
	// NOTE: we hold the lock even it's doing HTTP call to avoid sending multiple requests
	cachedOIDCIssuerProvidersMutex.Lock()
	defer cachedOIDCIssuerProvidersMutex.Unlock()

	if cached, ok := cachedOIDCIssuerProviders[issuerURL]; ok {
		// another goroutine has already constructed the provider
		return cached, nil
	}

	// NOTE: we start a root context here to allow background remote key set refresh
	ctx := context.Background()
	ctx = withRetryableHttpClient(ctx, issuerGetRetryCount)
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		// failed in this attempt, let other attempts retry
		return nil, errors.Wrap(err, "failed to create provider for azure")
	}

	cachedOIDCIssuerProviders[issuerURL] = provider

	return provider, nil
}

// New is called per authentication request
func New(ctx context.Context, opts Options) (auth.Interface, error) {
	c := &Authenticator{
		Options: opts,
	}

	if cachedAuthInfo == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if cachedAuthInfo == nil {
			// getAuthInfo gets the issuer, graph, and AAD endpoints from the Azure metadata endpoint
			// since it won't change at all, we can cache it
			authInfoVal, err := getAuthInfo(ctx, c.Environment, c.TenantID, c.HttpClientRetryCount, getMetadata)
			if err != nil {
				return nil, err
			}
			cachedAuthInfo = authInfoVal
		}
	}

	klog.V(3).Infof("Using issuer url: %v", cachedAuthInfo.Issuer)

	provider, err := getOIDCIssuerProvider(cachedAuthInfo.Issuer, c.HttpClientRetryCount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create provider for azure")
	}

	c.verifier = provider.Verifier(&oidc.Config{SkipClientIDCheck: !opts.VerifyClientID, ClientID: opts.ClientID})
	if opts.EnablePOP {
		c.popTokenVerifier = NewPoPVerifier(c.POPTokenHostname, c.PoPTokenValidityDuration)
	}

	switch opts.AuthMode {
	case ClientCredentialAuthMode:
		c.graphClient, err = graph.New(c.ClientID, c.ClientSecret, c.TenantID, c.UseGroupUID, cachedAuthInfo.AADEndpoint, cachedAuthInfo.MSGraphHost)
	case ARCAuthMode:
		c.graphClient, err = graph.NewWithARC(c.ClientID, c.ResourceId, c.TenantID, c.AzureRegion)
	case OBOAuthMode:
		c.graphClient, err = graph.NewWithOBO(c.ClientID, c.ClientSecret, c.TenantID, cachedAuthInfo.AADEndpoint, cachedAuthInfo.MSGraphHost)
	case AKSAuthMode:
		c.graphClient, err = graph.NewWithAKS(c.AKSTokenURL, c.TenantID, cachedAuthInfo.MSGraphHost)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ms graph client")
	}
	return c, nil
}

// makeRetryableHttpClient creates an HTTP client which attempts the request
// (1 + retryCount) times and has a 3 second timeout per attempt.
func makeRetryableHttpClient(retryCount int) retryablehttp.Client {
	// Copy the default HTTP client so we can set a timeout.
	// (It uses the same transport since the pointer gets copied)
	httpClient := *httpclient.DefaultHTTPClient
	httpClient.Timeout = 3 * time.Second

	// Attempt the request up to 3 times
	return retryablehttp.Client{
		HTTPClient:   &httpClient,
		RetryWaitMin: 500 * time.Millisecond,
		RetryWaitMax: 2 * time.Second,
		RetryMax:     retryCount, // initial + retryCount retries = (1 + retryCount) attempts
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
		Logger:       log.Default(),
	}
}

// withRetryableHttpClient sets the oauth2.HTTPClient key of the context to an
// *http.Client made from makeRetryableHttpClient.
// Some of the libraries we use will take the client out of the context via
// oauth2.HTTPClient and use it, so this way we can add retries to external code.
func withRetryableHttpClient(ctx context.Context, retryCount int) context.Context {
	retryClient := makeRetryableHttpClient(retryCount)
	return context.WithValue(ctx, oauth2.HTTPClient, retryClient.StandardClient())
}

type metadataJSON struct {
	Issuer      string `json:"issuer"`
	MsgraphHost string `json:"msgraph_host"`
}

// https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-convert-app-to-be-multi-tenant
func getMetadata(ctx context.Context, aadEndpoint, tenantID string, retryCount int) (*metadataJSON, error) {
	metadataURL := aadEndpoint + tenantID + "/.well-known/openid-configuration"
	retryClient := makeRetryableHttpClient(retryCount)

	request, err := retryablehttp.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := retryClient.Do(request.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get metadata url failed with status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
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

func (s Authenticator) Check(ctx context.Context, token string) (*authv1.UserInfo, error) {
	var err error

	if s.EnablePOP {
		token, err = s.popTokenVerifier.ValidatePopToken(token)
		if err != nil {
			return nil, errors.Wrap(err, "failed to verify pop token")
		}
	}

	ctx = withRetryableHttpClient(ctx, s.HttpClientRetryCount)
	idToken, err := s.verifier.Verify(ctx, token)
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

	if s.Options.ResolveGroupMembershipOnlyOnOverageClaim {
		groups, skipGraphAPI, err := getGroupsAndCheckOverage(claims)
		if err != nil {
			return nil, fmt.Errorf("error in getGroupsAndCheckOverage: %s", err)
		}
		if skipGraphAPI {
			resp.Groups = groups
			return resp, nil
		}
	}
	if !s.Options.SkipGroupMembershipResolution {
		if err := s.graphClient.RefreshToken(ctx, token); err != nil {
			return nil, err
		}
		resp.Groups, err = s.graphClient.GetGroups(ctx, resp.Username, token)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get groups")
		}

	}
	return resp, nil
}

// getGroupsAndCheckOverage will extract groups when groups claim is already present
// it also checks overage indicator and returns
// true: groups claim is present or overage indicator is not present. no need to call graph api
// false: there is a need to call graph api to get group membership
//
// overage indicator:
// https://docs.microsoft.com/en-us/azure/active-directory/develop/access-tokens#payload-claims
//
// ...
//
//	"_claim_names": {
//	  "groups": "src1"
//	},
//
//	"_claim_sources": {
//	  "src1": {
//	    "endpoint": "[Graph Url to get this user's group membership from]"
//	  }
//	},
//
// ...
func getGroupsAndCheckOverage(claims claims) ([]string, bool, error) {
	if c, ok := claims["groups"]; ok {
		var groups []string
		if err := marshalGenericTo(c, &groups); err != nil {
			return nil, true, err
		}
		return groups, true, nil
	}
	// we will short circuit when overage indicator is not present
	claimNames, ok := claims["_claim_names"]
	if !ok {
		return nil, true, nil
	}
	claimToSource := map[string]string{}
	if err := marshalGenericTo(claimNames, &claimToSource); err != nil {
		return nil, true, err
	}
	claimSources, ok := claims["_claim_sources"]
	if !ok {
		// it is not expected to have _claim_names but no _claim_sources
		// it will never get to this point because idToken.Verify()
		// already maps claim sources to claim names.
		// however, there is no interface to expose these resolved distributed claims
		return nil, true, errors.New("no _claim_sources is found")
	}
	var sources map[string]struct {
		Endpoint string `json:"endpoint"`
	}
	if err := marshalGenericTo(claimSources, &sources); err != nil {
		return nil, true, err
	}

	src, ok := claimToSource["groups"]
	if !ok {
		// no overage indicator present
		return nil, true, nil
	}
	ep, ok := sources[src]
	if !ok {
		// it will never get to this point because idToken.Verify()
		// already maps claim sources to claim names.
		// however, there is no interface to expose these resolved distributed claims
		return nil, true, fmt.Errorf("%s is missing in _claim_sources", src)
	}
	if ep.Endpoint == "" {
		// may not be a distributed token
		return nil, true, nil
	}

	// return true to proceed to call graph api
	return nil, false, nil
}

func marshalGenericTo(src interface{}, dst interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// GetClaims returns a Claims object
func getClaims(token *oidc.IDToken) (claims, error) {
	c := claims{}
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

	useroid, _ := c.string(userObjectIDClaim)

	return &authv1.UserInfo{
		Username: username,
		Extra:    map[string]authv1.ExtraValue{"oid": {useroid}},
	}, nil
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

type getMetadataFunc = func(context.Context, string, string, int) (*metadataJSON, error)

func getAuthInfo(ctx context.Context, environment, tenantID string, retryCount int, getMetadata getMetadataFunc) (*authInfo, error) {
	var err error
	env := azure.PublicCloud
	if environment != "" {
		env, err = azure.EnvironmentFromName(environment)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse environment for azure")
		}
	}

	metadata, err := getMetadata(ctx, env.ActiveDirectoryEndpoint, tenantID, retryCount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metadata for azure")
	}

	msgraphHost := metadata.MsgraphHost
	if strings.EqualFold(azure.USGovernmentCloud.Name, environment) {
		msgraphHost = "graph.microsoft.us"
	}

	return &authInfo{
		AADEndpoint: env.ActiveDirectoryEndpoint,
		MSGraphHost: msgraphHost,
		Issuer:      metadata.Issuer,
	}, nil
}

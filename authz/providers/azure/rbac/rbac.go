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
package rbac

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	v "github.com/appscode/go/version"
	auth "github.com/appscode/guard/auth/providers/azure"
	"github.com/appscode/guard/auth/providers/azure/graph"
	"github.com/appscode/guard/authz"
	authzOpts "github.com/appscode/guard/authz/providers/azure/options"

	"github.com/golang/glog"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	authzv1 "k8s.io/api/authorization/v1"
)

const (
	managedClusters           = "Microsoft.ContainerService/managedClusters"
	connectedClusters         = "Microsoft.Kubernetes/connectedClusters"
	checkAccessPath           = "/providers/Microsoft.Authorization/checkaccess"
	checkAccessAPIVersion     = "2018-09-01-preview"
	remainingSubReadARMHeader = "x-ms-ratelimit-remaining-subscription-reads"
	expiryDelta               = 60 * time.Second
)

type AuthzInfo struct {
	AADEndpoint string
	ARMEndPoint string
}

type void struct{}

// AccessInfo allows you to check user access from MS RBAC
type AccessInfo struct {
	headers   http.Header
	client    *http.Client
	expiresAt time.Time
	// These allow us to mock out the URL for testing
	apiURL *url.URL

	tokenProvider                  graph.TokenProvider
	clusterType                    string
	azureResourceId                string
	armCallLimit                   int
	skipCheck                      map[string]void
	retrieveGroupMemberships       bool
	skipAuthzForNonAADUsers        bool
	allowNonResDiscoveryPathAccess bool
	lock                           sync.RWMutex
}

func getClusterType(clsType string) string {
	switch clsType {
	case authzOpts.ARCAuthzMode:
		return connectedClusters
	case authzOpts.AKSAuthzMode:
		return managedClusters
	default:
		return ""
	}
}

func newAccessInfo(tokenProvider graph.TokenProvider, rbacURL *url.URL, opts authzOpts.Options) (*AccessInfo, error) {
	u := &AccessInfo{
		client: http.DefaultClient,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
			"User-Agent":   []string{fmt.Sprintf("%s-%s-%s-%s", v.Version.Platform, v.Version.GoVersion, v.Version.Version, opts.AuthzMode)},
		},
		apiURL:                         rbacURL,
		tokenProvider:                  tokenProvider,
		azureResourceId:                opts.ResourceId,
		armCallLimit:                   opts.ARMCallLimit,
		retrieveGroupMemberships:       opts.AuthzResolveGroupMemberships,
		skipAuthzForNonAADUsers:        opts.SkipAuthzForNonAADUsers,
		allowNonResDiscoveryPathAccess: opts.AllowNonResDiscoveryPathAccess,
	}

	u.skipCheck = make(map[string]void, len(opts.SkipAuthzCheck))
	var member void
	for _, s := range opts.SkipAuthzCheck {
		u.skipCheck[strings.ToLower(s)] = member
	}

	u.clusterType = getClusterType(opts.AuthzMode)
	u.lock = sync.RWMutex{}

	return u, nil
}

func New(opts authzOpts.Options, authopts auth.Options, authzInfo *AuthzInfo) (*AccessInfo, error) {
	rbacURL, err := url.Parse(authzInfo.ARMEndPoint)
	if err != nil {
		return nil, err
	}

	var tokenProvider graph.TokenProvider
	switch opts.AuthzMode {
	case authzOpts.ARCAuthzMode:
		tokenProvider = graph.NewClientCredentialTokenProvider(authopts.ClientID, authopts.ClientSecret,
			fmt.Sprintf("%s%s/oauth2/v2.0/token", authzInfo.AADEndpoint, authopts.TenantID),
			fmt.Sprintf("%s.default", authzInfo.ARMEndPoint))
	case authzOpts.AKSAuthzMode:
		tokenProvider = graph.NewAKSTokenProvider(opts.AKSAuthzTokenURL, authopts.TenantID)
	}

	return newAccessInfo(tokenProvider, rbacURL, opts)
}

func (a *AccessInfo) RefreshToken() error {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.IsTokenExpired() {
		resp, err := a.tokenProvider.Acquire("")
		if err != nil {
			glog.Errorf("%s failed to refresh token : %s", a.tokenProvider.Name(), err.Error())
			return errors.Wrap(err, "failed to refresh rbac token")
		}

		// Set the authorization headers for future requests
		a.headers.Set("Authorization", fmt.Sprintf("Bearer %s", resp.Token))
		expIn := time.Duration(resp.Expires) * time.Second
		a.expiresAt = time.Now().Add(expIn - expiryDelta)
		glog.Infof("Token refreshed successfully on %s. Expire at:%s", time.Now(), a.expiresAt)
	}

	return nil
}

func (a *AccessInfo) IsTokenExpired() bool {
	return a.expiresAt.Before(time.Now())
}

func (a *AccessInfo) ShouldSkipAuthzCheckForNonAADUsers() bool {
	return a.skipAuthzForNonAADUsers
}

func (a *AccessInfo) GetResultFromCache(request *authzv1.SubjectAccessReviewSpec, store authz.Store) (bool, bool) {
	var result bool
	key := getResultCacheKey(request)
	glog.V(10).Infof("Cache search for key: %s", key)
	found, _ := store.Get(key, &result)
	return found, result
}

func (a *AccessInfo) SkipAuthzCheck(request *authzv1.SubjectAccessReviewSpec) bool {
	if a.clusterType == connectedClusters {
		_, ok := a.skipCheck[strings.ToLower(request.User)]
		return ok
	}
	return false
}

func (a *AccessInfo) SetResultInCache(request *authzv1.SubjectAccessReviewSpec, result bool, store authz.Store) error {
	key := getResultCacheKey(request)
	glog.V(10).Infof("Cache set for key: %s, value: %t", key, result)
	return store.Set(key, result)
}

func (a *AccessInfo) AllowNonResPathDiscoveryAccess(request *authzv1.SubjectAccessReviewSpec) bool {
	if request.NonResourceAttributes != nil && a.allowNonResDiscoveryPathAccess && strings.EqualFold(request.NonResourceAttributes.Verb, "get") {
		path := strings.ToLower(request.NonResourceAttributes.Path)
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/openapi") || strings.HasPrefix(path, "/version") || strings.HasPrefix(path, "/healthz") {
			return true
		}
	}
	return false
}

func (a *AccessInfo) setReqHeaders(req *http.Request) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	// Set the auth headers for the request
	if req.Header == nil {
		req.Header = make(http.Header)
	}

	for k, value := range a.headers {
		req.Header[k] = value
	}
}

func (a *AccessInfo) CheckAccess(request *authzv1.SubjectAccessReviewSpec) (*authzv1.SubjectAccessReviewStatus, error) {
	checkAccessBody, err := prepareCheckAccessRequestBody(request, a.clusterType, a.azureResourceId, a.retrieveGroupMemberships)

	if err != nil {
		return nil, errors.Wrap(err, "error in preparing check access request")
	}

	checkAccessURL := *a.apiURL
	// Append the path for azure cluster resource id
	checkAccessURL.Path = path.Join(checkAccessURL.Path, a.azureResourceId)
	exist, nameSpaceString := getNameSpaceScope(request)
	if exist {
		checkAccessURL.Path = path.Join(checkAccessURL.Path, nameSpaceString)
	}

	checkAccessURL.Path = path.Join(checkAccessURL.Path, checkAccessPath)
	params := url.Values{}
	params.Add("api-version", checkAccessAPIVersion)
	checkAccessURL.RawQuery = params.Encode()

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(checkAccessBody); err != nil {
		return nil, errors.Wrap(err, "error encoding check access request")
	}

	if glog.V(10) {
		binaryData, _ := json.MarshalIndent(checkAccessBody, "", "    ")
		glog.V(10).Infof("checkAccessURI:%s", checkAccessURL.String())
		glog.V(10).Infof("binary data:%s", binaryData)
	}

	req, err := http.NewRequest(http.MethodPost, checkAccessURL.String(), buf)
	if err != nil {
		return nil, errors.Wrap(err, "error creating check access request")
	}

	a.setReqHeaders(req)

	if glog.V(10) {
		cmd, _ := http2curl.GetCurlCommand(req)
		glog.V(10).Infoln(cmd)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error in check access request execution")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error in reading response body")
	}

	defer resp.Body.Close()
	glog.V(10).Infof("checkaccess response: %s, Configured ARM call limit: %d", string(data), a.armCallLimit)
	if resp.StatusCode != http.StatusOK {
		glog.Errorf("error in check access response. error code: %d, response: %s", resp.StatusCode, string(data))
		if resp.StatusCode == http.StatusTooManyRequests {
			glog.V(10).Infoln("Closing idle TCP connections.")
			a.client.CloseIdleConnections()
			// TODO: add prom metrics for this scenario
		}
		return nil, errors.Errorf("request %s failed with status code: %d and response: %s", req.URL.Path, resp.StatusCode, string(data))
	} else {
		remaining := resp.Header.Get(remainingSubReadARMHeader)
		glog.Infof("Remaining request count in ARM instance:%s", remaining)
		count, _ := strconv.Atoi(remaining)
		if count < a.armCallLimit {
			if glog.V(10) {
				glog.V(10).Infoln("Closing idle TCP connections.")
			}
			// Usually ARM connections are cached by destination ip and port
			// By closing the idle connection, a new request will use different port which
			// will connect to different ARM instance of the region to ensure there is no ARM throttling
			a.client.CloseIdleConnections()
		}
	}

	// Decode response and prepare k8s response
	return ConvertCheckAccessResponse(data)
}

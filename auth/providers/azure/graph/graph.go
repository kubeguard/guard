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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/klog/v2"
)

// These are the base URL endpoints for MS graph
var (
	json                  = jsoniter.ConfigCompatibleWithStandardLibrary
	expandedGroupsPerCall = 500
	idtypClaim            = "idtyp"

	getMemberGroupsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "guard_azure_graph_failure_total",
		Help: "Azure graph getMemberGroups call failed.",
	})

	getOBORegionalEndpoint = getOBORegionalEndpointFunc
)

const (
	expiryDelta            = 60 * time.Second
	getMemberGroupsTimeout = 23 * time.Second
	getterName             = "ms-graph"
	arcAuthMode            = "arc"
	arcOboEndpointFormat   = "https://%s.obo.arc.azure.%s:8084%s/getMemberGroups?api-version=v1"
)

// UserInfo allows you to get user data from MS Graph
type UserInfo struct {
	headers http.Header
	client  *http.Client
	expires time.Time
	// These allow us to mock out the URL for testing
	apiURL *url.URL

	groupsPerCall int
	useGroupUID   bool

	tokenProvider TokenProvider
	authMode      string
	tenantID      string
	resourceID    string
	region        string
	lock          sync.RWMutex
}

func (u *UserInfo) getGroupIDs(ctx context.Context, userPrincipal string) ([]string, error) {
	// Create a new request for finding the user.
	// Shallow copy of the base API URL
	userSearchURL := *u.apiURL
	// Append the path for the member list
	userSearchURL.Path = path.Join(userSearchURL.Path, fmt.Sprintf("/users/%s/getMemberGroups", userPrincipal))

	// The body being sent makes sure that all groups are returned, not just security groups
	req, err := http.NewRequest(http.MethodPost, userSearchURL.String(), strings.NewReader(`{"securityEnabledOnly": false}`))
	if err != nil {
		return nil, errors.Wrap(err, "error creating group IDs request")
	}
	// Set the auth headers for the request
	req.Header = u.headers

	resp, err := u.client.Do(req.WithContext(ctx))
	if err != nil {
		getMemberGroupsFailed.Inc()
		return nil, errors.Wrap(err, "error listing users")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("request %s failed with status code: %d and response: %s", req.URL.Path, resp.StatusCode, string(data))
	}

	// Decode the group response
	objects := ObjectList{}
	err = json.NewDecoder(resp.Body).Decode(&objects)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode response for request %s", req.URL.Path)
	}
	return objects.Value, nil
}

func (u *UserInfo) getExpandedGroups(ctx context.Context, ids []string) (*GroupList, error) {
	// Encode the ids into the request body
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(ObjectQuery{
		IDs:   ids,
		Types: []string{"group"},
	})
	if err != nil {
		return nil, errors.Wrap(err, "error encoding body")
	}

	// Set up the request
	// Shallow copy of the base API URL
	groupObjectsURL := *u.apiURL
	// Append the path for the group expansion
	groupObjectsURL.Path = path.Join(groupObjectsURL.Path, "/directoryObjects/getByIds")
	req, err := http.NewRequest(http.MethodPost, groupObjectsURL.String(), body)
	if err != nil {
		return nil, errors.Wrap(err, "error creating group expansion request")
	}
	// Set the auth headers
	req.Header = u.headers

	if klog.V(10).Enabled() {
		cmd, _ := http2curl.GetCurlCommand(req)
		klog.V(10).Infoln(cmd)
	}

	resp, err := u.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "error expanding groups")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("request %s failed with status code: %d and response: %s", req.URL.Path, resp.StatusCode, string(data))
	}

	// Decode the response
	groups := &GroupList{}
	err = json.NewDecoder(resp.Body).Decode(groups)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode response for request %s", req.URL.Path)
	}
	return groups, nil
}

// GetMemberGroupsUsingARCOboService gets a list of all groups that the given user principal is part of using the ARC OBO service
func (u *UserInfo) getMemberGroupsUsingARCOboService(ctx context.Context, accessToken string) ([]string, error) {
	reqBody := struct {
		TenantID    string `json:"tenantID"`
		AccessToken string `json:"accessToken"`
	}{
		TenantID:    u.tenantID,
		AccessToken: accessToken,
	}

	claims := jwt.MapClaims{}
	// ParseUnverfied
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(accessToken, claims)
	if err != nil {
		if parsedToken == nil {
			return nil, errors.Wrap(err, "Error while parsing accessToken for validation, token is nil")
		}
		return nil, errors.Wrap(err, "Error while parsing accessToken for validation")
	}

	// the arc obo service does not support getting groups for applications
	if claims[idtypClaim] != nil {
		return nil, errors.New("Overage claim (users with more than 200 group membership) for SPN is currently not supported. For troubleshooting, please refer to aka.ms/overageclaimtroubleshoot")
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
		return nil, errors.Wrap(err, "failed to encode token request")
	}
	endpoint, err := getOBORegionalEndpoint(u.region, u.resourceID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create getMemberGroups request")
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create getMemberGroups request")
	}
	// Set the auth headers
	req.Header = u.headers

	correlationID := uuid.New()
	req.Header.Set("x-ms-correlation-request-id", correlationID.String())
	getMemberGroupsCtx, cancel := context.WithTimeout(context.Background(), getMemberGroupsTimeout)
	defer cancel()
	klog.V(5).Infof("Sending getMemberGroups request with correlationID: %s", correlationID.String())
	// use default httpclient without retries first
	client := *httpclient.DefaultHTTPClient
	resp, err := client.Do(req.WithContext(getMemberGroupsCtx))
	if err != nil {
		klog.V(5).Infof("CorrelationID: %s, Error: Failed to fetch group info using getMemberGroups: %s", correlationID.String(), err.Error())
		return nil, errors.Errorf("CorrelationID: %s, Error: Failed to fetch group info using getMemberGroups", correlationID.String())
	}

	// use retryable client only for unavailable and gatewaytimeout errors
	if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusGatewayTimeout {
		resp, err = u.client.Do(req.WithContext(ctx))
		if err != nil {
			klog.V(5).Infof("CorrelationID: %s, Error: Failed to fetch group info using getMemberGroups on retries: %s", correlationID.String(), err.Error())
			return nil, errors.Errorf("CorrelationID: %s, Error: Failed to fetch group info using getMemberGroups on retries", correlationID.String())
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		klog.V(5).Infof("CorrelationID: %s, Error: Failed to fetch group info with status code: %d and response: %s", correlationID.String(), resp.StatusCode, string(data))
		return nil, errors.Errorf("CorrelationID: %s, Error: Failed to fetch group info with status code: %d", correlationID.String(), resp.StatusCode)
	}

	groupResponse := struct {
		Value []string `json:"value"`
	}{}
	// Decode the response
	var groupIDs []string
	err = json.NewDecoder(resp.Body).Decode(&groupResponse)
	if err != nil {
		return nil, errors.Wrapf(err, "CorrelationID: %s, Error: Failed to decode response for request %s", correlationID.String(), req.URL.Path)
	}

	// Extract out the Group objects into a list of strings
	for i := 0; i < len(groupResponse.Value); i++ {
		groupIDs = append(groupIDs, groupResponse.Value[i])
	}

	totalGroups := len(groupIDs)
	klog.V(10).Infof("No of groups returned by OBO service: %d", totalGroups)

	return groupIDs, nil
}

func getOBORegionalEndpointFunc(location string, resourceID string) (string, error) {
	var suffix string

	if strings.HasPrefix(location, "usgov") || strings.HasPrefix(location, "usdod") {
		suffix = "us"
	} else if strings.HasPrefix(location, "china") {
		suffix = "cn"
	} else {
		suffix = "com"
	}

	if !strings.HasPrefix(resourceID, "/") {
		resourceID = "/" + resourceID
	}
	return fmt.Sprintf(arcOboEndpointFormat, location, suffix, resourceID), nil
}

func (u *UserInfo) RefreshToken(ctx context.Context, token string) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.isTokenExpired() {
		resp, err := u.tokenProvider.Acquire(ctx, token)
		if err != nil {
			return errors.Errorf("%s: failed to refresh token: %s", u.tokenProvider.Name(), err)
		}
		// Set the authorization headers for future requests
		u.headers.Set("Authorization", fmt.Sprintf("Bearer %s", resp.Token))
		expIn := time.Duration(resp.Expires) * time.Second
		u.expires = time.Now().Add(expIn - expiryDelta)
		klog.Infof("Token refreshed successfully on %s. Expire at:%s", time.Now(), u.expires)
	}

	return nil
}

func (u *UserInfo) isTokenExpired() bool {
	return u.expires.Before(time.Now())
}

// GetGroups gets a list of all groups that the given user principal is part of
// Generally in federated directories the email address is the userPrincipalName
func (u *UserInfo) GetGroups(ctx context.Context, userPrincipal string, token string) ([]string, error) {
	// use arc obo service to get groups if authn mode is arc
	if u.authMode == arcAuthMode {
		groupIds, err := u.getMemberGroupsUsingARCOboService(ctx, token)
		if err != nil {
			return nil, err
		}
		return groupIds, nil
	}
	// Get the group IDs for the user
	groupIDs, err := u.getGroupIDs(ctx, userPrincipal)
	if err != nil {
		return nil, err
	}

	if u.useGroupUID {
		return groupIDs, nil
	}

	totalGroups := len(groupIDs)
	klog.V(10).Infof("totalGroups: %d", totalGroups)

	groupNames := make([]string, 0, totalGroups)
	for i := 0; i < totalGroups; i += u.groupsPerCall {
		startIndex := i
		endIndex := min(i+u.groupsPerCall, totalGroups)
		klog.V(10).Infof("Getting group names for IDs between startIndex: %d and endIndex: %d", startIndex, endIndex)

		// Expand the group IDs
		groups, err := u.getExpandedGroups(ctx, groupIDs[startIndex:endIndex])
		if err != nil {
			return nil, err
		}
		// Extract out the Group objects into a list of strings
		for i := 0; i < len(groups.Value); i++ {
			groupNames = append(groupNames, groups.Value[i].Name)
		}
	}

	return groupNames, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Name returns the name of this getter
func (u *UserInfo) Name() string {
	return getterName
}

// newUserInfo returns a UserInfo object
func newUserInfo(tokenProvider TokenProvider, graphURL *url.URL, useGroupUID bool) (*UserInfo, error) {
	u := &UserInfo{
		client: httpclient.DefaultHTTPClient,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		apiURL:        graphURL,
		groupsPerCall: expandedGroupsPerCall,
		useGroupUID:   useGroupUID,
		tokenProvider: tokenProvider,
	}

	u.lock = sync.RWMutex{}

	return u, nil
}

// New returns a new UserInfo object
func New(clientID, clientSecret, tenantID string, useGroupUID bool, aadEndpoint, msgraphHost string) (*UserInfo, error) {
	graphEndpoint := "https://" + msgraphHost + "/"
	graphURL, _ := url.Parse(graphEndpoint + "v1.0")

	tokenProvider := NewClientCredentialTokenProvider(clientID, clientSecret,
		fmt.Sprintf("%s%s/oauth2/v2.0/token", aadEndpoint, tenantID),
		fmt.Sprintf("https://%s/.default", msgraphHost))

	return newUserInfo(tokenProvider, graphURL, useGroupUID)
}

// NewWithOBO returns a new UserInfo object
func NewWithOBO(clientID, clientSecret, tenantID string, aadEndpoint, msgraphHost string) (*UserInfo, error) {
	graphEndpoint := "https://" + msgraphHost + "/"
	graphURL, _ := url.Parse(graphEndpoint + "v1.0")

	tokenProvider := NewOBOTokenProvider(clientID, clientSecret,
		fmt.Sprintf("%s%s/oauth2/v2.0/token", aadEndpoint, tenantID),
		fmt.Sprintf("https://%s/.default", msgraphHost))

	return newUserInfo(tokenProvider, graphURL, true)
}

// NewWithAKS returns a new UserInfo object used in AKS
func NewWithAKS(tokenURL, tenantID, msgraphHost string) (*UserInfo, error) {
	graphEndpoint := "https://" + msgraphHost + "/"
	graphURL, _ := url.Parse(graphEndpoint + "v1.0")

	tokenProvider := NewAKSTokenProvider(tokenURL, tenantID)

	return newUserInfo(tokenProvider, graphURL, true)
}

// NewWithARC returns a new UserInfo object used in ARC
func NewWithARC(msiAudience, resourceId, tenantId, region string) (*UserInfo, error) {
	graphURL, _ := url.Parse("")
	tokenProvider := NewMSITokenProvider(msiAudience, MSIEndpointForARC)

	userInfo, err := newUserInfo(tokenProvider, graphURL, false)
	if err != nil {
		return nil, err
	}
	userInfo.tenantID = tenantId
	userInfo.resourceID = resourceId
	userInfo.region = region
	userInfo.authMode = arcAuthMode
	return userInfo, nil
}

func TestUserInfo(clientID, clientSecret, loginUrl, apiUrl string, useGroupUID bool) (*UserInfo, error) {
	parsedApi, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}
	u := &UserInfo{
		client: httpclient.DefaultHTTPClient,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		apiURL:        parsedApi,
		groupsPerCall: expandedGroupsPerCall,
		useGroupUID:   useGroupUID,
	}
	u.tokenProvider = NewClientCredentialTokenProvider(clientID, clientSecret, loginUrl, "")
	if err != nil {
		return nil, err
	}

	return u, nil
}

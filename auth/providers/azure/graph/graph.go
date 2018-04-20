package graph

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/json-iterator/go"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
)

// These are the base URL endpoints for MS graph
var (
	baseAPIURL, _ = url.Parse("https://graph.microsoft.com/v1.0")
	loginURL      = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
	json          = jsoniter.ConfigCompatibleWithStandardLibrary
)

const (
	graphDefaultScope = "https://graph.microsoft.com/.default" // This requests that a token use all of its default scopes
	graphGrantType    = "client_credentials"                   // The only grant type supported for this login flow
	expiryDelta       = 60 * time.Second
	getterName        = "ms-graph"
)

// UserInfo allows you to get user data from MS Graph
type UserInfo struct {
	headers      http.Header
	client       *http.Client
	clientID     string
	clientSecret string
	expires      time.Time
	// These allow us to mock out the URL for testing
	apiURL   *url.URL
	loginURL *url.URL
}

func (u *UserInfo) login() error {
	// Perform the login with the proper credentials
	// Put together the form data
	form := url.Values{}
	form.Set("client_id", u.clientID)
	form.Set("client_secret", u.clientSecret)
	form.Set("scope", graphDefaultScope)
	form.Set("grant_type", graphGrantType)

	req, err := http.NewRequest(http.MethodPost, u.loginURL.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return errors.Wrap(err, "error creating login request")
	}
	cmd, _ := http2curl.GetCurlCommand(req)
	fmt.Println("Request: ", cmd)

	resp, err := u.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error performing login")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf("request %s failed with status code: %d and response: %s", req.URL.Path, resp.StatusCode, string(data))
	}
	// Decode the response
	var authResp = &AuthResponse{}
	err = json.NewDecoder(resp.Body).Decode(authResp)
	if err != nil {
		return errors.Wrapf(err, "failed to decode response for request %s", &req.URL.Path)
	}

	// Set the authorization headers for future requests
	u.headers.Set("Authorization", fmt.Sprintf("Bearer %s", authResp.Token))
	expIn := time.Duration(authResp.Expires) * time.Second
	u.expires = time.Now().Add(expIn - expiryDelta)
	return nil
}

func (u *UserInfo) isExpired() bool {
	return time.Now().After(u.expires)
}

func (u *UserInfo) getGroupIDs(userPrincipal string) ([]string, error) {
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

	cmd, _ := http2curl.GetCurlCommand(req)
	fmt.Println("Request: ", cmd)

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error listing users")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Errorf("request %s failed with status code: %d and response: %s", req.URL.Path, resp.StatusCode, string(data))
	}

	// Decode the group response
	var objects = ObjectList{}
	err = json.NewDecoder(resp.Body).Decode(&objects)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode response for request %s", &req.URL.Path)
	}
	return objects.Value, nil
}

func (u *UserInfo) getExpandedGroups(ids []string) (*GroupList, error) {
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

	cmd, _ := http2curl.GetCurlCommand(req)
	fmt.Println("Request: ", cmd)

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error expanding groups")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Errorf("request %s failed with status code: %d and response: %s", req.URL.Path, resp.StatusCode, string(data))
	}

	// Decode the response
	var groups = &GroupList{}
	err = json.NewDecoder(resp.Body).Decode(groups)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode response for request %s", &req.URL.Path)
	}
	return groups, nil
}

// GetGroups gets a list of all groups that the given user principal is part of
// Generally in federated directories the email address is the userPrincipalName
func (u *UserInfo) GetGroups(userPrincipal string) ([]string, error) {
	// Make sure things are logged in before continuing
	if err := u.login(); err != nil {
		return nil, err
	}

	// Get the group IDs for the user
	objIDs, err := u.getGroupIDs(userPrincipal)
	if err != nil {
		return nil, err
	}

	// Expand the group IDs
	groups, err := u.getExpandedGroups(objIDs)
	if err != nil {
		return nil, err
	}

	// Extract out the Group objects into a list of strings
	var finalList = make([]string, len(groups.Value))
	for i := 0; i < len(groups.Value); i++ {
		finalList[i] = groups.Value[i].Name
	}
	return finalList, nil
}

// Name returns the name of this getter
func (u *UserInfo) Name() string {
	return getterName
}

// New returns a new UserInfo object
func New(clientID, clientSecret, tenantName string) (*UserInfo, error) {
	parsedLogin, err := url.Parse(fmt.Sprintf(loginURL, tenantName))
	if err != nil {
		return nil, err
	}
	u := &UserInfo{
		client: http.DefaultClient,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		apiURL:       baseAPIURL,
		loginURL:     parsedLogin,
		clientID:     clientID,
		clientSecret: clientSecret,
	}

	return u, nil
}

func NewUserInfo(clientID, clientSecret, tenantName, loginUrl, apiUrl string) (*UserInfo, error) {
	parsedLogin, err := url.Parse(loginUrl)
	if err != nil {
		return nil, err
	}
	parsedApi, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}
	u := &UserInfo{
		client: http.DefaultClient,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		apiURL:       parsedApi,
		loginURL:     parsedLogin,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
	err = u.login()
	if err != nil {
		return nil, err
	}

	return u, nil
}

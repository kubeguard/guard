package graph

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
)

type aksTokenRefresher struct {
	name     string
	client   *http.Client
	tokenURL string
	tenantID string
}

// NewAKSTokenRefresher returns a TokenRefresher that implements On-Behalf-Of flow using AKS first party service
func NewAKSTokenRefresher(tokenURL, tenantID string) TokenRefresher {
	return &aksTokenRefresher{
		name:     "AKSTokenRefresher",
		client:   &http.Client{},
		tokenURL: tokenURL,
		tenantID: tenantID,
	}
}

func (u *aksTokenRefresher) Name() string { return u.name }

func (u *aksTokenRefresher) Refresh(token string) (AuthResponse, error) {
	var authResp = AuthResponse{}
	tokenReq := struct {
		TenantID    string `json:"tenantID,omitempty"`
		AccessToken string `json:"accessToken,omitempty"`
	}{
		TenantID:    u.tenantID,
		AccessToken: token,
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
	if glog.V(10) {
		cmd, _ := http2curl.GetCurlCommand(req)
		glog.V(10).Infoln(cmd)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return authResp, errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return authResp, errors.Errorf("request failed with status code: %d and response: %s", resp.StatusCode, string(data))
	}
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return authResp, errors.Wrapf(err, "failed to decode response")
	}

	return authResp, nil
}

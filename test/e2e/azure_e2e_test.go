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

package e2e_test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	azureauth "go.kubeguard.dev/guard/auth/providers/azure"
	"go.kubeguard.dev/guard/test/e2e/framework"

	"github.com/golang-jwt/jwt/v4"
	"gomodules.xyz/cert"
	authv1 "k8s.io/api/authentication/v1"
)

const (
	azureE2EClientCommonName = "azure-e2e-client"
	azureE2ERequestTimeout   = 10 * time.Second
)

type expectedAzureTokenReviewUser struct {
	Username string
	ObjectID string
}

func expectedAzureUserInfoFromAccessToken(rawAccessToken string) (expectedAzureTokenReviewUser, error) {
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(rawAccessToken, jwt.MapClaims{})
	if err != nil {
		return expectedAzureTokenReviewUser{}, fmt.Errorf("failed to parse Azure access token: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return expectedAzureTokenReviewUser{}, fmt.Errorf("Azure access token claims had unexpected type %T", parsedToken.Claims)
	}

	objectID, _ := stringClaim(claims, "oid")
	if objectID == "" {
		return expectedAzureTokenReviewUser{}, fmt.Errorf("Azure access token does not contain an oid claim")
	}

	username, _ := stringClaim(claims, "upn")
	if username == "" {
		username = objectID
	}

	return expectedAzureTokenReviewUser{Username: username, ObjectID: objectID}, nil
}

func expectedAzureUserInfoFromPoPToken(rawPoPToken, hostName string) (expectedAzureTokenReviewUser, error) {
	innerAccessToken, err := azureauth.NewPoPVerifier(hostName, 15*time.Minute).ValidatePopToken(rawPoPToken)
	if err != nil {
		return expectedAzureTokenReviewUser{}, fmt.Errorf("failed to extract inner access token from PoP token: %w", err)
	}

	return expectedAzureUserInfoFromAccessToken(innerAccessToken)
}

func sendTokenReviewRequest(
	localPort uint16,
	serverName, org, rawToken string,
	invocation *framework.Invocation,
) (*authv1.TokenReview, int, error) {
	clientCrt, clientKey, err := invocation.CertStore.NewClientCertPairBytes(
		cert.AltNames{DNSNames: []string{azureE2EClientCommonName}},
		org,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create client certificate: %w", err)
	}

	tlsCert, err := tls.X509KeyPair(clientCrt, clientKey)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse client certificate: %w", err)
	}

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(invocation.CertStore.CACertBytes()) {
		return nil, 0, fmt.Errorf("failed to append Guard CA certificate")
	}

	httpClient := &http.Client{
		Timeout: azureE2ERequestTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{tlsCert},
				RootCAs:      rootCAs,
				ServerName:   serverName,
				MinVersion:   tls.VersionTLS12,
			},
		},
	}

	reviewBody, err := json.Marshal(authv1.TokenReview{Spec: authv1.TokenReviewSpec{Token: rawToken}})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal tokenreview request: %w", err)
	}

	requestURL := fmt.Sprintf("https://127.0.0.1:%d/tokenreviews", localPort)
	req, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(reviewBody))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create tokenreview request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute tokenreview request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read tokenreview response: %w", err)
	}

	var review authv1.TokenReview
	if err := json.Unmarshal(body, &review); err != nil {
		return nil, resp.StatusCode, fmt.Errorf(
			"failed to decode tokenreview response: %w: %s",
			err,
			strings.TrimSpace(string(body)),
		)
	}

	if resp.StatusCode != http.StatusOK {
		return &review, resp.StatusCode, fmt.Errorf(
			"tokenreview request failed with status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(review.Status.Error),
		)
	}

	return &review, resp.StatusCode, nil
}

func stringClaim(claims jwt.MapClaims, name string) (string, bool) {
	value, ok := claims[name]
	if !ok {
		return "", false
	}

	stringValue, ok := value.(string)
	if !ok {
		return "", false
	}

	stringValue = strings.TrimSpace(stringValue)
	if stringValue == "" {
		return "", false
	}

	return stringValue, true
}


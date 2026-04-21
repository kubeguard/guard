package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	azureutils "go.kubeguard.dev/guard/util/azure"

	"github.com/coreos/go-oidc"
	"github.com/hashicorp/go-retryablehttp"
)

type AccessTokenVerifier interface {
	Verify(ctx context.Context, rawAccessToken string) (VerifiedAccessToken, error)
}

type VerifiedAccessToken interface {
	Claims() (claims, error)
}

var (
	_ AccessTokenVerifier = (*OIDCAccessTokenVerifier)(nil)
	_ AccessTokenVerifier = (*EntraSDKTokenVerifier)(nil)
	_ VerifiedAccessToken = (*oidcVerifiedAccessToken)(nil)
	_ VerifiedAccessToken = (*staticClaimsToken)(nil)
)

type OIDCAccessTokenVerifier struct {
	Verifier *oidc.IDTokenVerifier
}

func (o *OIDCAccessTokenVerifier) Verify(ctx context.Context, rawAccessToken string) (VerifiedAccessToken, error) {
	token, err := o.Verifier.Verify(ctx, rawAccessToken)
	if err != nil {
		return nil, err
	}

	return &oidcVerifiedAccessToken{token: token}, nil
}

type oidcVerifiedAccessToken struct {
	token *oidc.IDToken
}

func (t *oidcVerifiedAccessToken) Claims() (claims, error) {
	if t.token == nil {
		return nil, fmt.Errorf("claims not set")
	}

	parsedClaims := claims{}
	if err := t.token.Claims(&parsedClaims); err != nil {
		return nil, err
	}

	return parsedClaims, nil
}

type EntraSDKTokenVerifier struct {
	baseURL              *url.URL
	clientID             string
	verifyClientID       bool
	httpClientRetryCount int
}

func newEntraSDKTokenVerifier(rawBaseURL, clientID string, verifyClientID bool, httpClientRetryCount int) (*EntraSDKTokenVerifier, error) {
	parsedBaseURL, err := parseEntraSDKBaseURL(rawBaseURL)
	if err != nil {
		return nil, err
	}

	return &EntraSDKTokenVerifier{
		baseURL:              parsedBaseURL,
		clientID:             clientID,
		verifyClientID:       verifyClientID,
		httpClientRetryCount: httpClientRetryCount,
	}, nil
}

func parseEntraSDKBaseURL(rawBaseURL string) (*url.URL, error) {
	trimmed := strings.TrimSpace(rawBaseURL)
	if trimmed == "" {
		return nil, fmt.Errorf("Entra SDK endpoint is empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid Entra SDK endpoint: %w", err)
	}

	if parsed.Path != "" && parsed.Path != "/" {
		return nil, fmt.Errorf("Entra SDK endpoint must be a base URL")
	}

	// Only keep the scheme and host, and discard any user info, path, query,
	// or fragment details to ensure a clean base URL.
	return &url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host,
	}, nil
}

func (e *EntraSDKTokenVerifier) Verify(ctx context.Context, rawAccessToken string) (VerifiedAccessToken, error) {
	if e.baseURL == nil {
		return nil, fmt.Errorf("Entra SDK verifier is not initialized")
	}

	validateURL := *e.baseURL
	validateURL.Path = "/Validate"

	retryClient := azureutils.MakeRetryableHttpClient(ctx, e.httpClientRetryCount)
	req, err := retryablehttp.NewRequest(http.MethodGet, validateURL.String(), nil)
	if err != nil {
		return nil, err
	}
	if hostHeader := entraSDKRequestHost(e.baseURL); hostHeader != "" {
		req.Host = hostHeader
	}
	req.Header.Set("Authorization", "Bearer "+rawAccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := retryClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to call Entra SDK validate endpoint: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Entra SDK validate response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseEntraSDKError(resp.StatusCode, body)
	}

	var validated entraSDKValidateResponse
	if err := json.Unmarshal(body, &validated); err != nil {
		return nil, fmt.Errorf("failed to decode Entra SDK validate response: %w", err)
	}
	if len(validated.Claims) == 0 {
		return nil, fmt.Errorf("Entra SDK validate response did not include claims")
	}

	// Verify the audience directly, just in case
	if e.verifyClientID {
		aud, ok := validated.Claims["aud"]
		if !ok {
			return nil, fmt.Errorf("claim aud not found")
		}
		if !audienceContains(aud, e.clientID) {
			return nil, fmt.Errorf("expected audience %q got %v", e.clientID, aud)
		}
	}

	return &staticClaimsToken{parsedClaims: validated.Claims}, nil
}

func entraSDKRequestHost(baseURL *url.URL) string {
	if baseURL == nil {
		return ""
	}

	hostname := strings.TrimSpace(baseURL.Hostname())
	if hostname == "" {
		return ""
	}

	if strings.EqualFold(hostname, "localhost") {
		return "localhost"
	}

	if ip := net.ParseIP(hostname); ip != nil && ip.IsLoopback() {
		return "localhost"
	}

	return hostname
}

type staticClaimsToken struct {
	parsedClaims claims
}

func (t *staticClaimsToken) Claims() (claims, error) {
	return t.parsedClaims, nil
}

type entraSDKValidateResponse struct {
	Protocol string                 `json:"protocol"`
	Token    string                 `json:"token"`
	Claims   map[string]interface{} `json:"claims"`
}

type entraSDKErrorResponse struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

// parseEntraSDKError attempts to extract meaningful error information from the Entra SDK response body.
func parseEntraSDKError(statusCode int, body []byte) error {
	var sdkErr entraSDKErrorResponse
	if err := json.Unmarshal(body, &sdkErr); err == nil {
		switch {
		case sdkErr.Detail != "":
			if isExpectedEntraSDKValidationStatus(statusCode) {
				return buildTokenValidationError(sdkErr.Detail)
			}
			return fmt.Errorf("Entra SDK validate request failed with status %d: %s", statusCode, sdkErr.Detail)
		case sdkErr.Title != "":
			if isExpectedEntraSDKValidationStatus(statusCode) {
				return buildTokenValidationError(sdkErr.Title)
			}
			return fmt.Errorf("Entra SDK validate request failed with status %d: %s", statusCode, sdkErr.Title)
		}
	}

	trimmedBody := strings.TrimSpace(string(body))
	if isExpectedEntraSDKValidationStatus(statusCode) {
		return buildTokenValidationError(trimmedBody)
	}

	if trimmedBody == "" {
		return fmt.Errorf("Entra SDK validate request failed with status %d", statusCode)
	}

	return fmt.Errorf("Entra SDK validate request failed with status %d: %s", statusCode, trimmedBody)
}

func isExpectedEntraSDKValidationStatus(statusCode int) bool {
	return statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError
}

func buildTokenValidationError(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("token failed validation")
	}

	return fmt.Errorf("token failed validation: %s", reason)
}

// audienceContains checks if the audience claim contains the expected audience.
// The audience claim can be a string or an array of strings.
func audienceContains(audienceClaim interface{}, expectedAudience string) bool {
	switch aud := audienceClaim.(type) {
	case string:
		return aud == expectedAudience
	case []string:
		for _, value := range aud {
			if value == expectedAudience {
				return true
			}
		}
	case []interface{}:
		for _, value := range aud {
			if audValue, ok := value.(string); ok && audValue == expectedAudience {
				return true
			}
		}
	}

	return false
}

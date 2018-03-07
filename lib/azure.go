package lib

import (
	"context"
	"fmt"
	"net/http"

	"github.com/appscode/go/log"
	"github.com/appscode/guard/lib/graph"
	oidc "github.com/coreos/go-oidc"
	"github.com/spf13/pflag"
	auth "k8s.io/api/authentication/v1"
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
	azureIssuerURL     = "https://sts.windows.net/"
	azureUsernameClaim = "upn"
	azureGroupClaim    = "groups"
)

var (
	// ErrorClaimNotFound indicates the given key was not found in the claims
	ErrorClaimNotFound = fmt.Errorf("claim not found")
)

// claims represents a map of claims provided with a JWT
type claims map[string]interface{}

type AzureOptions struct {
	ClientID     string
	ClientSecret string
	TenantID     string
}

type AzureClient struct {
	AzureOptions
	graphClient *graph.UserInfo
	verifier    *oidc.IDTokenVerifier
	ctx         context.Context
}

func (s *AzureOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.ClientID, "azure.client-id", s.ClientID, "MS Graph application client ID to use")
	fs.StringVar(&s.ClientSecret, "azure.client-secret", s.ClientSecret, "MS Graph application client secret to use")
	fs.StringVar(&s.TenantID, "azure.tenant-id", s.TenantID, "MS Graph application tenant id to use")
}

func (s AzureOptions) ToArgs() []string {
	var args []string

	if s.ClientID != "" {
		args = append(args, fmt.Sprintf("--azure.client-id=%s", s.ClientID))
	}
	if s.ClientSecret != "" {
		args = append(args, fmt.Sprintf("--azure.client-secret=%s", s.ClientSecret))
	}
	if s.TenantID != "" {
		args = append(args, fmt.Sprintf("--azure.tenant-id=%s", s.TenantID))
	}

	return args
}

func NewAzureClient(opts AzureOptions) (*AzureClient, error) {
	c := &AzureClient{
		AzureOptions: opts,
		ctx:          context.Background(),
	}

	var err error
	provider, err := oidc.NewProvider(c.ctx, azureIssuerURL+c.TenantID+"/")
	if err != nil {
		return nil, fmt.Errorf("Failed to create provider for azure. Reason: %v.", err)
	}

	c.verifier = provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	c.graphClient, err = graph.New(c.ClientID, c.ClientSecret, c.TenantID)
	if err != nil {
		return nil, fmt.Errorf("Failed to create ms graph client. Reason %v.", err)
	}

	return c, nil
}

func (s AzureClient) checkAzure(token string) (auth.TokenReview, int) {
	idToken, err := s.verifier.Verify(s.ctx, token)
	if err != nil {
		return Error(fmt.Sprintf("Failed to verify token for azure. Reason: %v.", err)), http.StatusUnauthorized
	}

	claims, err := getClaims(idToken)
	if err != nil {
		return Error(fmt.Sprintf("Error parsing claims: %s", err)), http.StatusUnauthorized
	}

	finalReview, err := claims.ReviewFromClaims(azureUsernameClaim, azureGroupClaim)
	if err != nil {
		log.Infof("Failed to create TokenReview")
		return Error(fmt.Sprintf("Failed to create TokenReview. Reason %v.", err)), http.StatusBadRequest
	}

	finalReview.Status.User.Groups, err = s.graphClient.GetGroups(finalReview.Status.User.Username)
	if err != nil {
		log.Info("Failed to get groups")
		return Error(fmt.Sprintf("Failed to get groups. Reason %v.", err)), http.StatusBadRequest
	}
	return *finalReview, http.StatusOK
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
func (c claims) ReviewFromClaims(usernameClaim, groupsClaim string) (*auth.TokenReview, error) {
	var review = &auth.TokenReview{
		Status: auth.TokenReviewStatus{
			Authenticated: true,
		},
	}
	username, err := c.String(usernameClaim)
	if err != nil {
		if err == ErrorClaimNotFound {
			return nil, fmt.Errorf("username claim %s not found", usernameClaim)
		}
		return nil, fmt.Errorf("unable to get username claim: %s", err)
	}
	review.Status.User.Username = username

	//groups contain group-id
	groups, err := c.StringSlice(groupsClaim)
	if err != nil {
		if err == ErrorClaimNotFound {
			// Don't error out if groups claim is not present.
			// Just log a warning
			log.Infof("Groups is empty")
			groups = []string{}
		} else {
			return nil, fmt.Errorf("unable to get groups claim: %s", err)
		}
	}
	review.Status.User.Groups = groups
	return review, nil
}

func (c claims) hasKey(key string) bool {
	_, ok := c[key]
	return ok
}

// String gets a string value from claims given a key. Returns false if
// the key does not exist
func (c claims) String(key string) (string, error) {
	var resp string
	if !c.hasKey(key) {
		return "", ErrorClaimNotFound
	}
	if v, ok := c[key].(string); ok {
		resp = v
	} else { // Not a string type
		return "", fmt.Errorf("claim is not a string")
	}
	return resp, nil
}

// StringSlice gets a slice of strings from claims given a key. Returns false if
// the key does not exist
func (c claims) StringSlice(key string) ([]string, error) {
	var resp []string
	var intermediate []interface{}
	if !c.hasKey(key) {
		return nil, ErrorClaimNotFound
	}
	if val, ok := c[key].([]interface{}); ok {
		intermediate = val
	} else {
		return nil, fmt.Errorf("claim is not a slice")
	}
	// Initialize the slice to the same length as the intermediate slice. This saves
	// some steps with not having to append
	resp = make([]string, len(intermediate))
	// You can't type assert the whole slice as a type, so assert each element to make
	// sure it is a string
	for i := 0; i < len(resp); i++ {
		if strVal, ok := intermediate[i].(string); ok {
			resp[i] = strVal
		} else {
			return nil, fmt.Errorf("claim is not a slice of strings")
		}
	}
	return resp, nil
}

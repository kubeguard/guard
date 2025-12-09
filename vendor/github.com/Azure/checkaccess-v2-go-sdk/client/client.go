package client

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	"github.com/Azure/checkaccess-v2-go-sdk/client/internal/token"
)

// this asserts that &remotePDPClient{} would always implement RemotePDPClient
var _ RemotePDPClient = &remotePDPClient{}

// RemotePDPClient represents the Microsoft Remote PDP API Spec
type RemotePDPClient interface {
	CheckAccess(context.Context, AuthorizationRequest) (*AuthorizationDecisionResponse, error)
	CreateAuthorizationRequest(string, []string, string) (*AuthorizationRequest, error)
}

// remotePDPClient implements RemotePDPClient
type remotePDPClient struct {
	endpoint string
	pipeline runtime.Pipeline
}

// NewRemotePDPClient returns an implementation of RemotePDPClient
// endpoint - the fqdn of the regional specific endpoint of PDP
// scope - the oauth scope required by the PDP server
// cred - the credential of the client to call the PDP server
// ClientOptions - the optional settings for a client's pipeline.
func NewRemotePDPClient(endpoint, scope string, cred azcore.TokenCredential, clientOptions *azcore.ClientOptions) (*remotePDPClient, error) {
	if strings.TrimSpace(endpoint) == "" {
		return nil, fmt.Errorf("endpoint: %s is not valid, need a valid endpoint in creating client", endpoint)
	}
	if strings.TrimSpace(scope) == "" {
		return nil, fmt.Errorf("scope: %s is not valid, need a valid scope in creating client", scope)
	}
	if cred == nil {
		return nil, fmt.Errorf("need TokenCredential in creating client")
	}

	authPolicy := runtime.NewBearerTokenPolicy(cred, []string{scope}, nil)

	pipeline := runtime.NewPipeline(
		modulename,
		version,
		runtime.PipelineOptions{
			PerCall:  []policy.Policy{},
			PerRetry: []policy.Policy{authPolicy},
		},
		clientOptions,
	)

	return &remotePDPClient{endpoint, pipeline}, nil
}

// CheckAccess sends an Authorization query to the PDP server specified in the client
// ctx - the context to propagate
// authzReq - the actual AuthorizationRequest
func (r *remotePDPClient) CheckAccess(ctx context.Context, authzReq AuthorizationRequest) (*AuthorizationDecisionResponse, error) {
	req, err := runtime.NewRequest(ctx, http.MethodPost, r.endpoint)
	if err != nil {
		return nil, err
	}
	if err := runtime.MarshalAsJSON(req, authzReq); err != nil {
		return nil, err
	}

	res, err := r.pipeline.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, newCheckAccessError(res)
	}

	var accessDecision AuthorizationDecisionResponse
	if err := runtime.UnmarshalAsJSON(res, &accessDecision); err != nil {
		return nil, err
	}

	return &accessDecision, nil
}

// newCheckAccessError returns an error when non HTTP 200 response is returned.
func newCheckAccessError(r *http.Response) error {
	payload, err := runtime.Payload(r)
	if err != nil {
		return err
	}
	var checkAccessError CheckAccessErrorResponse
	err = json.Unmarshal(payload, &checkAccessError)
	if err != nil {
		return err
	}
	return &azcore.ResponseError{
		StatusCode:  r.StatusCode,
		RawResponse: r,
		ErrorCode:   fmt.Sprint(checkAccessError.StatusCode),
	}
}

// CreateAuthorizationRequest creates an AuthorizationRequest object
func (r *remotePDPClient) CreateAuthorizationRequest(resourceId string, actions []string, jwtToken string) (*AuthorizationRequest, error) {
	if strings.TrimSpace(jwtToken) == "" {
		return nil, fmt.Errorf("need token in creating AuthorizationRequest")
	}

	tokenClaims, err := token.ExtractClaims(jwtToken)
	if err != nil {
		return nil, fmt.Errorf("error while parse the token, err: %v", err)
	}

	subjectAttributes := SubjectAttributes{}
	subjectAttributes.ObjectId = tokenClaims.ObjectId

	if tokenClaims.ClaimNames != nil && len(tokenClaims.Groups) == 0 {
		subjectAttributes.ClaimName = GroupExpansion
	} else if tokenClaims.ClaimNames == nil && len(tokenClaims.Groups) > 0 {
		subjectAttributes.Groups = tokenClaims.Groups
	}

	actionInfos := []ActionInfo{}
	for _, action := range actions {
		actionInfos = append(actionInfos, ActionInfo{Id: action})
	}

	return &AuthorizationRequest{
		Subject: SubjectInfo{
			Attributes: subjectAttributes,
		},
		Actions: actionInfos,
		Resource: ResourceInfo{
			Id: resourceId,
		},
	}, nil
}

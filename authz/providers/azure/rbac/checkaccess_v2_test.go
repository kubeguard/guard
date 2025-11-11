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

/*
This file contains unit tests for CheckAccess v2 API implementation.

Test coverage includes:
- Response conversion (all allowed, partial denial, complete denial, empty)
- JWT token extraction from providers
- Data action conversion from SubjectAccessReviewSpec
- Batch processing with multiple batches (200 actions per batch)
- Error handling (request creation errors, PDP server errors)
- Primary authorization flow
- Fallback scenarios (managed namespace, fleet)

All tests use mock PDP client to avoid external dependencies.
*/

package rbac

import (
	"context"
	"errors"
	"testing"

	checkaccess "github.com/Azure/checkaccess-v2-go-sdk/client"
	"github.com/stretchr/testify/assert"
	"go.kubeguard.dev/guard/auth/providers/azure/graph"
	authzv1 "k8s.io/api/authorization/v1"
)

// mockPDPClient is a mock implementation of checkaccess.RemotePDPClient for testing
type mockPDPClient struct {
	createAuthzReqFunc func(resourceId string, actions []string, jwtToken string) (*checkaccess.AuthorizationRequest, error)
	checkAccessFunc    func(ctx context.Context, authzReq checkaccess.AuthorizationRequest) (*checkaccess.AuthorizationDecisionResponse, error)
}

func (m *mockPDPClient) CreateAuthorizationRequest(resourceId string, actions []string, jwtToken string) (*checkaccess.AuthorizationRequest, error) {
	if m.createAuthzReqFunc != nil {
		return m.createAuthzReqFunc(resourceId, actions, jwtToken)
	}
	return &checkaccess.AuthorizationRequest{}, nil
}

func (m *mockPDPClient) CheckAccess(ctx context.Context, authzReq checkaccess.AuthorizationRequest) (*checkaccess.AuthorizationDecisionResponse, error) {
	if m.checkAccessFunc != nil {
		return m.checkAccessFunc(ctx, authzReq)
	}
	return &checkaccess.AuthorizationDecisionResponse{}, nil
}

func TestConvertV2ResponseToStatus_AllAllowed(t *testing.T) {
	ctx := context.Background()
	decisions := []checkaccess.AuthorizationDecision{
		{
			ActionId:       "Microsoft.ContainerService/managedClusters/pods/read",
			AccessDecision: checkaccess.Allowed,
			RoleAssignment: checkaccess.RoleAssignment{
				Id:               "/subscriptions/sub/providers/Microsoft.Authorization/roleAssignments/123",
				RoleDefinitionId: "/subscriptions/sub/providers/Microsoft.Authorization/roleDefinitions/456",
			},
		},
		{
			ActionId:       "Microsoft.ContainerService/managedClusters/deployments/write",
			AccessDecision: checkaccess.Allowed,
		},
	}

	status := convertV2ResponseToStatus(ctx, decisions)

	assert.True(t, status.Allowed)
	assert.False(t, status.Denied)
	assert.Contains(t, status.Reason, "Access allowed by Azure RBAC")
	assert.Contains(t, status.Reason, "roleAssignments/123")
	assert.Contains(t, status.Reason, "roleDefinitions/456")
}

func TestConvertV2ResponseToStatus_OneDenied(t *testing.T) {
	ctx := context.Background()
	decisions := []checkaccess.AuthorizationDecision{
		{
			ActionId:       "Microsoft.ContainerService/managedClusters/pods/read",
			AccessDecision: checkaccess.Allowed,
		},
		{
			ActionId:       "Microsoft.ContainerService/managedClusters/pods/delete",
			AccessDecision: checkaccess.NotAllowed,
		},
	}

	status := convertV2ResponseToStatus(ctx, decisions)

	assert.False(t, status.Allowed)
	assert.True(t, status.Denied)
	assert.Contains(t, status.Reason, "Access denied for action")
	assert.Contains(t, status.Reason, "pods/delete")
}

func TestConvertV2ResponseToStatus_AllDenied(t *testing.T) {
	ctx := context.Background()
	decisions := []checkaccess.AuthorizationDecision{
		{
			ActionId:       "Microsoft.ContainerService/managedClusters/pods/delete",
			AccessDecision: checkaccess.Denied,
		},
	}

	status := convertV2ResponseToStatus(ctx, decisions)

	assert.False(t, status.Allowed)
	assert.True(t, status.Denied)
	assert.Contains(t, status.Reason, "Access denied for action")
}

func TestConvertV2ResponseToStatus_EmptyDecisions(t *testing.T) {
	ctx := context.Background()
	decisions := []checkaccess.AuthorizationDecision{}

	status := convertV2ResponseToStatus(ctx, decisions)

	assert.False(t, status.Allowed)
	assert.True(t, status.Denied)
	assert.Equal(t, AccessNotAllowedVerdict, status.Reason)
}

func TestGetJWTTokenFromProvider_Success(t *testing.T) {
	mockProvider := &mockTokenProvider{
		name: "test-provider",
		response: graph.AuthResponse{
			Token:     "test-jwt-token",
			ExpiresOn: 1234567890,
		},
	}

	accessInfo := &AccessInfo{
		tokenProvider: mockProvider,
	}

	ctx := context.Background()
	token, err := accessInfo.getJWTTokenFromProvider(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "test-jwt-token", token)
}

func TestGetJWTTokenFromProvider_Error(t *testing.T) {
	mockProvider := &mockTokenProvider{
		name: "test-provider",
		err:  errors.New("token acquisition failed"),
	}

	accessInfo := &AccessInfo{
		tokenProvider: mockProvider,
	}

	ctx := context.Background()
	token, err := accessInfo.getJWTTokenFromProvider(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to acquire JWT token")
	assert.Contains(t, err.Error(), "test-provider")
	assert.Empty(t, token)
}

func TestGetDataActionsV2_Success(t *testing.T) {
	ctx := context.Background()
	request := &authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace: "default",
			Verb:      "get",
			Resource:  "pods",
		},
	}

	actions, err := getDataActionsV2(ctx, request, managedClusters, false, false)

	assert.NoError(t, err)
	assert.NotEmpty(t, actions)
	assert.Contains(t, actions[0], "pods")
	assert.Contains(t, actions[0], "read")
}

func TestPerformCheckAccessV2_Success(t *testing.T) {
	mockClient := &mockPDPClient{
		createAuthzReqFunc: func(resourceId string, actions []string, jwtToken string) (*checkaccess.AuthorizationRequest, error) {
			return &checkaccess.AuthorizationRequest{}, nil
		},
		checkAccessFunc: func(ctx context.Context, authzReq checkaccess.AuthorizationRequest) (*checkaccess.AuthorizationDecisionResponse, error) {
			return &checkaccess.AuthorizationDecisionResponse{
				Value: []checkaccess.AuthorizationDecision{
					{
						ActionId:       "Microsoft.ContainerService/managedClusters/pods/read",
						AccessDecision: checkaccess.Allowed,
						RoleAssignment: checkaccess.RoleAssignment{
							Id:               "/roleAssignments/123",
							RoleDefinitionId: "/roleDefinitions/456",
						},
					},
				},
			}, nil
		},
	}

	accessInfo := &AccessInfo{
		pdpClient: mockClient,
	}

	ctx := context.Background()
	actions := []string{"Microsoft.ContainerService/managedClusters/pods/read"}
	status, err := accessInfo.performCheckAccessV2(ctx, "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster", actions, "jwt-token")

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Allowed)
	assert.False(t, status.Denied)
}

func TestPerformCheckAccessV2_BatchingMultipleBatches(t *testing.T) {
	// Create 250 actions to test batching (should create 2 batches: 200 + 50)
	actions := make([]string, 250)
	for i := 0; i < 250; i++ {
		actions[i] = "Microsoft.ContainerService/managedClusters/action"
	}

	callCount := 0
	mockClient := &mockPDPClient{
		createAuthzReqFunc: func(resourceId string, actions []string, jwtToken string) (*checkaccess.AuthorizationRequest, error) {
			// First batch should have 200, second should have 50
			switch callCount {
			case 0:
				assert.Equal(t, 200, len(actions))
			case 1:
				assert.Equal(t, 50, len(actions))
			}
			callCount++
			return &checkaccess.AuthorizationRequest{}, nil
		},
		checkAccessFunc: func(ctx context.Context, authzReq checkaccess.AuthorizationRequest) (*checkaccess.AuthorizationDecisionResponse, error) {
			// Return allowed for all
			decisions := make([]checkaccess.AuthorizationDecision, 1)
			decisions[0] = checkaccess.AuthorizationDecision{
				ActionId:       "action",
				AccessDecision: checkaccess.Allowed,
				RoleAssignment: checkaccess.RoleAssignment{Id: "id"},
			}
			return &checkaccess.AuthorizationDecisionResponse{Value: decisions}, nil
		},
	}

	accessInfo := &AccessInfo{
		pdpClient: mockClient,
	}

	ctx := context.Background()
	status, err := accessInfo.performCheckAccessV2(ctx, "/resource", actions, "jwt")

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, 2, callCount, "Should make 2 batched calls")
	assert.True(t, status.Allowed)
}

func TestPerformCheckAccessV2_CreateRequestError(t *testing.T) {
	mockClient := &mockPDPClient{
		createAuthzReqFunc: func(resourceId string, actions []string, jwtToken string) (*checkaccess.AuthorizationRequest, error) {
			return nil, errors.New("failed to create request")
		},
	}

	accessInfo := &AccessInfo{
		pdpClient: mockClient,
	}

	ctx := context.Background()
	status, err := accessInfo.performCheckAccessV2(ctx, "/resource", []string{"action"}, "jwt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create v2 authorization request")
	assert.Nil(t, status)
}

func TestPerformCheckAccessV2_CheckAccessError(t *testing.T) {
	mockClient := &mockPDPClient{
		createAuthzReqFunc: func(resourceId string, actions []string, jwtToken string) (*checkaccess.AuthorizationRequest, error) {
			return &checkaccess.AuthorizationRequest{}, nil
		},
		checkAccessFunc: func(ctx context.Context, authzReq checkaccess.AuthorizationRequest) (*checkaccess.AuthorizationDecisionResponse, error) {
			return nil, errors.New("PDP server error")
		},
	}

	accessInfo := &AccessInfo{
		pdpClient: mockClient,
	}

	ctx := context.Background()
	status, err := accessInfo.performCheckAccessV2(ctx, "/resource", []string{"action"}, "jwt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CheckAccess v2 batch failed")
	assert.Contains(t, err.Error(), "PDP server error")
	assert.Nil(t, status)
}

func TestCheckAccessV2_PrimaryAllowed(t *testing.T) {
	mockClient := &mockPDPClient{
		createAuthzReqFunc: func(resourceId string, actions []string, jwtToken string) (*checkaccess.AuthorizationRequest, error) {
			return &checkaccess.AuthorizationRequest{}, nil
		},
		checkAccessFunc: func(ctx context.Context, authzReq checkaccess.AuthorizationRequest) (*checkaccess.AuthorizationDecisionResponse, error) {
			return &checkaccess.AuthorizationDecisionResponse{
				Value: []checkaccess.AuthorizationDecision{
					{
						ActionId:       "action",
						AccessDecision: checkaccess.Allowed,
						RoleAssignment: checkaccess.RoleAssignment{Id: "id"},
					},
				},
			}, nil
		},
	}

	mockProvider := &mockTokenProvider{
		response: graph.AuthResponse{Token: "jwt-token"},
	}

	accessInfo := &AccessInfo{
		pdpClient:                       mockClient,
		tokenProvider:                   mockProvider,
		clusterType:                     managedClusters,
		azureResourceId:                 "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster",
		useNamespaceResourceScopeFormat: false,
	}

	ctx := context.Background()
	request := &authzv1.SubjectAccessReviewSpec{
		User: "test@example.com",
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace: "default",
			Verb:      "get",
			Resource:  "pods",
		},
	}

	status, err := accessInfo.checkAccessV2(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Allowed)
	assert.False(t, status.Denied)
}

func TestCheckAccessV2_FallbackToManagedNamespace(t *testing.T) {
	callCount := 0
	mockClient := &mockPDPClient{
		createAuthzReqFunc: func(resourceId string, actions []string, jwtToken string) (*checkaccess.AuthorizationRequest, error) {
			return &checkaccess.AuthorizationRequest{}, nil
		},
		checkAccessFunc: func(ctx context.Context, authzReq checkaccess.AuthorizationRequest) (*checkaccess.AuthorizationDecisionResponse, error) {
			callCount++
			decision := checkaccess.NotAllowed
			// First call (primary) returns denied, second call (managed namespace) returns allowed
			if callCount == 2 {
				decision = checkaccess.Allowed
			}
			return &checkaccess.AuthorizationDecisionResponse{
				Value: []checkaccess.AuthorizationDecision{
					{
						ActionId:       "action",
						AccessDecision: decision,
						RoleAssignment: checkaccess.RoleAssignment{Id: "id"},
					},
				},
			}, nil
		},
	}

	mockProvider := &mockTokenProvider{
		response: graph.AuthResponse{Token: "jwt-token"},
	}

	accessInfo := &AccessInfo{
		pdpClient:                              mockClient,
		tokenProvider:                          mockProvider,
		clusterType:                            managedClusters,
		azureResourceId:                        "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster",
		useManagedNamespaceResourceScopeFormat: true,
		useNamespaceResourceScopeFormat:        false,
	}

	ctx := context.Background()
	request := &authzv1.SubjectAccessReviewSpec{
		User: "test@example.com",
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace: "default",
			Verb:      "get",
			Resource:  "pods",
		},
	}

	status, err := accessInfo.checkAccessV2(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Allowed)
	assert.Equal(t, 2, callCount, "Should make 2 calls: primary + managed namespace")
}

func TestCheckAccessV2_FallbackToFleet(t *testing.T) {
	callCount := 0
	mockClient := &mockPDPClient{
		createAuthzReqFunc: func(resourceId string, actions []string, jwtToken string) (*checkaccess.AuthorizationRequest, error) {
			return &checkaccess.AuthorizationRequest{}, nil
		},
		checkAccessFunc: func(ctx context.Context, authzReq checkaccess.AuthorizationRequest) (*checkaccess.AuthorizationDecisionResponse, error) {
			callCount++
			decision := checkaccess.NotAllowed
			// Third call (fleet) returns allowed
			if callCount == 3 {
				decision = checkaccess.Allowed
			}
			return &checkaccess.AuthorizationDecisionResponse{
				Value: []checkaccess.AuthorizationDecision{
					{
						ActionId:       "action",
						AccessDecision: decision,
						RoleAssignment: checkaccess.RoleAssignment{Id: "id"},
					},
				},
			}, nil
		},
	}

	mockProvider := &mockTokenProvider{
		response: graph.AuthResponse{Token: "jwt-token"},
	}

	accessInfo := &AccessInfo{
		pdpClient:                              mockClient,
		tokenProvider:                          mockProvider,
		clusterType:                            managedClusters,
		azureResourceId:                        "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster",
		useManagedNamespaceResourceScopeFormat: true,
		fleetManagerResourceId:                 "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/fleets/fleet",
	}

	ctx := context.Background()
	request := &authzv1.SubjectAccessReviewSpec{
		User: "test@example.com",
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace: "default",
			Verb:      "get",
			Resource:  "pods",
		},
	}

	status, err := accessInfo.checkAccessV2(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Allowed)
	assert.Equal(t, 3, callCount, "Should make 3 calls: primary + managed namespace + fleet")
}

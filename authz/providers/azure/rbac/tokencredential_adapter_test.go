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
	"context"
	"errors"
	"testing"
	"time"

	"go.kubeguard.dev/guard/auth/providers/azure/graph"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/stretchr/testify/assert"
)

// mockTokenProvider is a mock implementation of graph.TokenProvider for testing
type mockTokenProvider struct {
	name     string
	response graph.AuthResponse
	err      error
}

func (m *mockTokenProvider) Name() string {
	return m.name
}

func (m *mockTokenProvider) Acquire(ctx context.Context, token string) (graph.AuthResponse, error) {
	if m.err != nil {
		return graph.AuthResponse{}, m.err
	}
	return m.response, nil
}

func TestTokenProviderAdapter_GetToken_Success(t *testing.T) {
	// Arrange
	expectedToken := "test-access-token"
	expectedExpiresOn := time.Now().Add(1 * time.Hour).Unix()

	mockProvider := &mockTokenProvider{
		name: "mock-provider",
		response: graph.AuthResponse{
			Token:     expectedToken,
			ExpiresOn: int(expectedExpiresOn),
			TokenType: "Bearer",
		},
	}

	adapter := newTokenProviderAdapter(mockProvider, "https://management.azure.com/.default")

	// Act
	ctx := context.Background()
	token, err := adapter.GetToken(ctx, policy.TokenRequestOptions{})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token.Token)
	assert.Equal(t, time.Unix(expectedExpiresOn, 0), token.ExpiresOn)
}

func TestTokenProviderAdapter_GetToken_ProviderError(t *testing.T) {
	// Arrange
	expectedError := errors.New("token acquisition failed")
	mockProvider := &mockTokenProvider{
		name: "mock-provider",
		err:  expectedError,
	}

	adapter := newTokenProviderAdapter(mockProvider, "https://management.azure.com/.default")

	// Act
	ctx := context.Background()
	token, err := adapter.GetToken(ctx, policy.TokenRequestOptions{})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to acquire token from provider mock-provider")
	assert.Contains(t, err.Error(), "token acquisition failed")
	assert.Empty(t, token.Token)
}

func TestTokenProviderAdapter_GetToken_ExpiryConversion(t *testing.T) {
	// Arrange
	// Test that Unix timestamp is correctly converted to time.Time
	specificTime := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	expectedExpiresOn := specificTime.Unix()

	mockProvider := &mockTokenProvider{
		name: "mock-provider",
		response: graph.AuthResponse{
			Token:     "token",
			ExpiresOn: int(expectedExpiresOn),
		},
	}

	adapter := newTokenProviderAdapter(mockProvider, "test-scope")

	// Act
	ctx := context.Background()
	token, err := adapter.GetToken(ctx, policy.TokenRequestOptions{})

	// Assert
	assert.NoError(t, err)
	// Compare Unix timestamps to avoid timezone issues
	assert.Equal(t, expectedExpiresOn, token.ExpiresOn.Unix())
}

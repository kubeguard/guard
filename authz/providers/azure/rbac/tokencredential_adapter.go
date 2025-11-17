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
	"fmt"
	"time"

	"go.kubeguard.dev/guard/auth/providers/azure/graph"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// tokenProviderAdapter adapts a graph.TokenProvider to azcore.TokenCredential interface
// for use with the CheckAccess v2 SDK.
type tokenProviderAdapter struct {
	provider graph.TokenProvider
	scope    string
}

// newTokenProviderAdapter creates a new adapter that wraps a graph.TokenProvider
// and implements azcore.TokenCredential interface.
func newTokenProviderAdapter(provider graph.TokenProvider, scope string) azcore.TokenCredential {
	return &tokenProviderAdapter{
		provider: provider,
		scope:    scope,
	}
}

// GetToken implements azcore.TokenCredential interface by adapting the graph.TokenProvider.Acquire method.
func (a *tokenProviderAdapter) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	// Call the underlying token provider
	authResp, err := a.provider.Acquire(ctx, "")
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to acquire token from provider %s: %w", a.provider.Name(), err)
	}

	// Convert AuthResponse to azcore.AccessToken
	expiresOn := time.Unix(int64(authResp.ExpiresOn), 0)
	return azcore.AccessToken{
		Token:     authResp.Token,
		ExpiresOn: expiresOn,
	}, nil
}

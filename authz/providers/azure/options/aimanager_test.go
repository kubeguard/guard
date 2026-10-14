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

package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAIManagerID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string
	}{
		// Valid AI Manager IDs
		{
			name:    "valid AI Manager ID",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/aiManagers/my-aim",
			wantErr: false,
		},
		{
			name:    "valid AI Manager ID with hyphens",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/aiManagers/my-aim-123",
			wantErr: false,
		},
		{
			name:    "valid AI Manager ID case-insensitive resource type",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/microsoft.containerservice/aimanagers/my-aim",
			wantErr: false,
		},

		// Invalid resource types
		{
			name:    "not an AI Manager resource - AKS cluster",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/managedClusters/my-cluster",
			wantErr: true,
			errMsg:  "is not an AI Manager resource",
		},
		{
			name:    "not an AI Manager resource - fleet",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet",
			wantErr: true,
			errMsg:  "is not an AI Manager resource",
		},
		{
			name:    "not an AI Manager resource - storage account",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.Storage/storageAccounts/mystorage",
			wantErr: true,
			errMsg:  "is not an AI Manager resource",
		},

		// Malformed resource IDs
		{
			name:    "malformed resource ID - missing subscription",
			id:      "/resourceGroups/my-rg/providers/Microsoft.ContainerService/aiManagers/my-aim",
			wantErr: true,
		},
		{
			name:    "malformed resource ID - invalid format",
			id:      "not-a-resource-id",
			wantErr: true,
		},
		{
			name:    "empty resource ID",
			id:      "",
			wantErr: true,
		},
		{
			name:    "malformed resource ID - missing providers",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/Microsoft.ContainerService/aiManagers/my-aim",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAIManagerID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

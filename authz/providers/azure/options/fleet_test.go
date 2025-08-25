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

func TestValidateFleetID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string
	}{
		// Valid fleet IDs
		{
			name:    "valid fleet ID",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet",
			wantErr: false,
		},
		{
			name:    "valid fleet ID with single character name",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg/providers/Microsoft.ContainerService/fleets/f",
			wantErr: false,
		},
		{
			name:    "valid fleet ID with hyphens",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet-123",
			wantErr: false,
		},
		{
			name:    "valid fleet ID with maximum length name (63 chars)",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxy0",
			wantErr: false,
		},
		{
			name:    "valid fleet ID with numbers",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/fleet123",
			wantErr: false,
		},

		// Invalid resource types
		{
			name:    "not a fleet resource - AKS cluster",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/managedClusters/my-cluster",
			wantErr: true,
			errMsg:  "is not a fleet resource",
		},
		{
			name:    "not a fleet resource - connected cluster",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.Kubernetes/connectedClusters/my-cluster",
			wantErr: true,
			errMsg:  "is not a fleet resource",
		},
		{
			name:    "not a fleet resource - storage account",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.Storage/storageAccounts/mystorage",
			wantErr: true,
			errMsg:  "is not a fleet resource",
		},

		// Invalid fleet names (regex validation)
		{
			name:    "fleet name with uppercase letters",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/MyFleet",
			wantErr: true,
			errMsg:  "does not have a valid fleet manager name",
		},
		{
			name:    "fleet name starting with hyphen",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/-fleet",
			wantErr: true,
			errMsg:  "does not have a valid fleet manager name",
		},
		{
			name:    "fleet name ending with hyphen",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/fleet-",
			wantErr: true,
			errMsg:  "does not have a valid fleet manager name",
		},
		{
			name:    "fleet name with underscore",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my_fleet",
			wantErr: true,
			errMsg:  "does not have a valid fleet manager name",
		},
		{
			name:    "fleet name with special characters",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my.fleet",
			wantErr: true,
			errMsg:  "does not have a valid fleet manager name",
		},
		{
			name:    "fleet name too long (64 chars)",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz012",
			wantErr: true,
			errMsg:  "does not have a valid fleet manager name",
		},
		{
			name:    "empty fleet name",
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/",
			wantErr: true,
		},

		// Malformed resource IDs
		{
			name:    "malformed resource ID - missing subscription",
			id:      "/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet",
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
			id:      "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/Microsoft.ContainerService/fleets/my-fleet",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFleetID(tt.id)
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

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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFleetManagerResourceID(t *testing.T) {
	tests := []struct {
		name                string
		runtimeConfigPath   string
		runtimeConfigReader func(string) ([]byte, error)
		expectedFound       bool
		expectedResourceID  string
	}{
		{
			name:               "empty runtime config path",
			runtimeConfigPath:  "",
			expectedFound:      false,
			expectedResourceID: "",
		},
		{
			name:              "valid fleet resource ID",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet"), nil
			},
			expectedFound:      true,
			expectedResourceID: "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet",
		},
		{
			name:              "valid fleet resource ID with special characters in resource group",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg-test-group/providers/Microsoft.ContainerService/fleets/fleet1"), nil
			},
			expectedFound:      true,
			expectedResourceID: "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg-test-group/providers/Microsoft.ContainerService/fleets/fleet1",
		},
		{
			name:              "valid fleet resource ID with hyphens in fleet name",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet-001"), nil
			},
			expectedFound:      true,
			expectedResourceID: "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet-001",
		},
		{
			name:              "valid fleet resource ID with whitespace",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("  /subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet  \n"), nil
			},
			expectedFound:      true,
			expectedResourceID: "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet",
		},
		{
			name:              "invalid fleet resource ID - wrong provider",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.Compute/fleets/my-fleet"), nil
			},
			expectedFound:      false,
			expectedResourceID: "",
		},
		{
			name:              "invalid fleet resource ID - wrong resource type",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/clusters/my-fleet"), nil
			},
			expectedFound:      false,
			expectedResourceID: "",
		},
		{
			name:              "invalid fleet resource ID - fleet name starts with hyphen",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/-my-fleet"), nil
			},
			expectedFound:      false,
			expectedResourceID: "",
		},
		{
			name:              "invalid fleet resource ID - fleet name ends with hyphen",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet-"), nil
			},
			expectedFound:      false,
			expectedResourceID: "",
		},
		{
			name:              "valid fleet resource ID - subscription starts with letter",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("/subscriptions/a2345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet"), nil
			},
			expectedFound:      true,
			expectedResourceID: "/subscriptions/a2345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet",
		},
		{
			name:              "invalid fleet resource ID - extra path segments",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet/extra"), nil
			},
			expectedFound:      false,
			expectedResourceID: "",
		},
		{
			name:              "file read error",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return nil, errors.New("file not found")
			},
			expectedFound:      false,
			expectedResourceID: "",
		},
		{
			name:              "empty file content",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte(""), nil
			},
			expectedFound:      false,
			expectedResourceID: "",
		},
		{
			name:              "only whitespace in file",
			runtimeConfigPath: "/test/config",
			runtimeConfigReader: func(path string) ([]byte, error) {
				return []byte("   \n\t  "), nil
			},
			expectedFound:      false,
			expectedResourceID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessInfo := &AccessInfo{
				runtimeConfigPath:   tt.runtimeConfigPath,
				runtimeConfigReader: tt.runtimeConfigReader,
			}

			found, resourceID := accessInfo.getFleetManagerResourceID()

			assert.Equal(t, tt.expectedFound, found, "Found flag mismatch")
			assert.Equal(t, tt.expectedResourceID, resourceID, "Resource ID mismatch")
		})
	}
}

func TestFleetResourceIDPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid simple fleet ID",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg1/providers/Microsoft.ContainerService/fleets/fleet1",
			expected: true,
		},
		{
			name:     "valid fleet ID with special chars in RG",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/my-rg-test-group/providers/Microsoft.ContainerService/fleets/fleet1",
			expected: true,
		},
		{
			name:     "valid fleet ID with hyphens in fleet name",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg1/providers/Microsoft.ContainerService/fleets/my-fleet-001",
			expected: true,
		},
		{
			name:     "case insensitive provider",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg1/providers/microsoft.containerservice/fleets/fleet1",
			expected: true,
		},
		{
			name:     "case insensitive fleet",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg1/providers/Microsoft.ContainerService/Fleets/fleet1",
			expected: true,
		},
		{
			name:     "invalid - wrong provider",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg1/providers/Microsoft.Compute/fleets/fleet1",
			expected: false,
		},
		{
			name:     "invalid - wrong resource type",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg1/providers/Microsoft.ContainerService/clusters/fleet1",
			expected: false,
		},
		{
			name:     "invalid - fleet name starts with hyphen",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg1/providers/Microsoft.ContainerService/fleets/-fleet1",
			expected: false,
		},
		{
			name:     "invalid - fleet name ends with hyphen",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg1/providers/Microsoft.ContainerService/fleets/fleet1-",
			expected: false,
		},
		{
			name:     "invalid - subscription starts with letter",
			input:    "/subscriptions/a2345678-1234-5678-9012-123456789012/resourceGroups/rg1/providers/Microsoft.ContainerService/fleets/fleet1",
			expected: true,
		},
		{
			name:     "invalid - extra path segments",
			input:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/rg1/providers/Microsoft.ContainerService/fleets/fleet1/extra",
			expected: false,
		},
		{
			name:     "invalid - fleet name is too long",
			input:    "/subscriptions/12345679-1234-1234-1234-123456789abc/resourceGroups/my-rg/providers/Microsoft.ContainerService/fleets/my-fleet/fleet1fleet1fleet1fleet1fleet1fleet1fleet1fleet1fleet1fleet1fleet1",
			expected: false,
		},
		{
			name:     "invalid - missing subscription",
			input:    "/resourceGroups/rg1/providers/Microsoft.ContainerService/fleets/fleet1",
			expected: false,
		},
		{
			name:     "invalid - empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid - only fleet name",
			input:    "fleet1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fleetResrcExp.MatchString(tt.input)
			assert.Equal(t, tt.expected, result, "Pattern match result mismatch for input: %s", tt.input)
		})
	}
}

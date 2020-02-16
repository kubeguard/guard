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

package providers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthProvidersHas(t *testing.T) {
	authCaseSensitive := AuthProviders{
		[]string{
			"azure",
			"github",
			"gitlab",
			"google",
			"ldap",
			"token-auth",
		},
	}

	authCaseInSensitive := AuthProviders{
		[]string{
			"AzUre",
			"GitHuB",
			"GitLAb",
			"GoOgLe",
			"LDap",
			"TokEn-auTh",
		},
	}

	type testProviderInfo struct {
		name         string
		expectedResp bool
	}

	testProvider := []testProviderInfo{
		{"azure", true},
		{"github", true},
		{"gitlab", true},
		{"google", true},
		{"ldap", true},
		{"token-auth", true},
		{"AzUre", true},
		{"GitHuB", true},
		{"GitLAb", true},
		{"GoOgLe", true},
		{"LDap", true},
		{"TokEn-auTh", true},
		{"AAure", false},
	}

	testData := []struct {
		testName      string
		authProviders AuthProviders
		testProviders []testProviderInfo
	}{
		{
			"auth provider are in small letter",
			authCaseSensitive,
			testProvider,
		},
		{
			"auth provider are in case insensitive letter",
			authCaseInSensitive,
			testProvider,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			for _, data := range test.testProviders {
				resp := test.authProviders.Has(data.name)
				assert.Equal(t, data.expectedResp, resp, fmt.Sprintf("testing for provider name %s for %v", data.name, test.authProviders))
			}
		})
	}
}

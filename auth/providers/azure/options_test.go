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

package azure

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

const (
	empty    = ""
	nonempty = "non-empty"
)

type optionFunc func(o Options) Options

type testInfo struct {
	testName    string
	opts        Options
	expectedErr []error
}

var validationErrorData = []struct {
	testName    string
	optsFunc    optionFunc
	expectedErr error
	allError    bool
}{
	{
		"azure.auth-mode is invalid",
		func(o Options) Options {
			o.AuthMode = empty
			return o
		},
		errors.New("invalid azure.auth-mode. valid value is either aks, obo, client-credential or passthrough"),
		true,
	},
	{
		"azure.client-secret must be non-empty",
		func(o Options) Options {
			o.ClientSecret = empty
			return o
		},
		errors.New("azure.client-secret must be non-empty"),
		true,
	},
	{
		"azure.client-id is empty when verify-clientID is true",
		func(o Options) Options {
			o.ClientID = empty
			o.VerifyClientID = true
			return o
		},
		errors.New("azure.client-id must be non-empty when azure.verify-clientID is set"),
		false,
	},
	{
		"azure.tenant-id is empty",
		func(o Options) Options {
			o.TenantID = empty
			return o
		},
		errors.New("azure.tenant-id must be non-empty"),
		true,
	},
	{
		"azure.aks-token-url is empty for aks auth mode",
		func(o Options) Options {
			o.AuthMode = AKSAuthMode
			return o
		},
		errors.New("azure.aks-token-url must be non-empty"),
		false,
	},
	{
		"azure.enable-pop is set without a hostname",
		func(o Options) Options {
			o.AuthMode = PassthroughAuthMode
			o.EnablePOP = true
			o.ResolveGroupMembershipOnlyOnOverageClaim = true
			o.SkipGroupMembershipResolution = true
			return o
		},
		errors.New("azure.pop-hostname must be non-empty when pop token is enabled"),
		false,
	},
	{
		"passthrough Auth Mode should force ResolveGroupMembershipOnlyOnOverageClaim to be set",
		func(o Options) Options {
			o.AuthMode = PassthroughAuthMode
			o.SkipGroupMembershipResolution = true
			return o
		},
		errors.New("azure.graph-call-on-overage-claim cannot be false when passthrough azure.auth-mode is used"),
		false,
	},
	{
		"passthrough Auth Mode should force SkipGroupMembershipResolution to be set",
		func(o Options) Options {
			o.AuthMode = PassthroughAuthMode
			o.ResolveGroupMembershipOnlyOnOverageClaim = true
			return o
		},
		errors.New("azure.skip-group-membership-resolution cannot be false when passthrough azure.auth-mode is used"),
		false,
	},
	{
		"azure.entra-sdk-url must be a valid URL",
		func(o Options) Options {
			o.EntraSDKURL = "://bad-url"
			return o
		},
		errors.New("azure.entra-sdk-url is not a valid URL: parse \"://bad-url\": missing protocol scheme"),
		false,
	},
	{
		"azure.entra-sdk-url must be a base URL",
		func(o Options) Options {
			o.EntraSDKURL = "http://localhost:8080/Validate"
			return o
		},
		errors.New("azure.entra-sdk-url must be a base URL"),
		false,
	},
	{
		"azure.client-id is required with Entra SDK URL",
		func(o Options) Options {
			o.EntraSDKURL = "http://localhost:8080"
			o.ClientID = empty
			return o
		},
		errors.New("azure.client-id must be non-empty when Entra SDK is enabled"),
		false,
	},
	{
		"azure.environment must resolve for Entra SDK",
		func(o Options) Options {
			o.EntraSDKURL = "http://localhost:8080"
			o.Environment = "definitely-not-a-real-cloud"
			return o
		},
		errors.New("failed to resolve Entra SDK Azure AD instance: autorest/azure: There is no cloud environment matching the name \"DEFINITELY-NOT-A-REAL-CLOUD\""),
		false,
	},
}

func getNonEmptyOptions() Options {
	return Options{
		ClientID:     nonempty,
		ClientSecret: nonempty,
		TenantID:     nonempty,
		AuthMode:     ClientCredentialAuthMode,
	}
}

func getEmptyOptions() Options {
	return Options{}
}

func getAllError() []error {
	var errs []error
	for _, d := range validationErrorData {
		if d.allError {
			errs = append(errs, d.expectedErr)
		}
	}
	return errs
}

func getTestDataForIndivitualError() []testInfo {
	test := []testInfo{}
	for _, d := range validationErrorData {
		test = append(test, testInfo{
			d.testName,
			d.optsFunc(getNonEmptyOptions()),
			[]error{d.expectedErr},
		})
	}

	return test
}

func TestOptionsValidate(t *testing.T) {
	testData := []testInfo{
		{
			"validation failed, all empty",
			getEmptyOptions(),
			getAllError(),
		},
		{
			"validation passed",
			getNonEmptyOptions(),
			nil,
		},
		{
			"validation passed with Entra SDK URL",
			func() Options {
				o := getNonEmptyOptions()
				o.EntraSDKURL = "http://localhost:8080"
				return o
			}(),
			nil,
		},
	}

	testData = append(testData, getTestDataForIndivitualError()...)

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			errs := test.opts.Validate()
			if test.expectedErr == nil {
				assert.Nil(t, errs)
			} else {
				if assert.NotNil(t, errs, "errors expected") {
					assert.EqualError(t, utilerrors.NewAggregate(errs), utilerrors.NewAggregate(test.expectedErr).Error())
				}
			}
		})
	}
}

func TestOptionsEntraSDKEnvVars(t *testing.T) {
	t.Run("returns expected env vars for public cloud", func(t *testing.T) {
		envVars, err := Options{
			ClientID: "client-id",
			TenantID: "tenant-id",
		}.EntraSDKEnvVars()

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"AzureAd__Instance", "AzureAd__TenantId", "AzureAd__ClientId", "AzureAd__Audience"}, []string{envVars[0].Name, envVars[1].Name, envVars[2].Name, envVars[3].Name})
			assert.Equal(t, "https://login.microsoftonline.com/", envVars[0].Value)
			assert.Equal(t, "tenant-id", envVars[1].Value)
			assert.Equal(t, "client-id", envVars[2].Value)
			assert.Equal(t, "client-id", envVars[3].Value)
		}
	})

	t.Run("uses the configured Azure environment", func(t *testing.T) {
		envVars, err := Options{
			Environment: "AzureChinaCloud",
			ClientID:    "client-id",
			TenantID:    "tenant-id",
		}.EntraSDKEnvVars()

		if assert.NoError(t, err) {
			assert.Equal(t, "https://login.chinacloudapi.cn/", envVars[0].Value)
		}
	})
}

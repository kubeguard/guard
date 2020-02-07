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

	aggregator "github.com/appscode/go/util/errors"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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

var (
	validationErrorData = []struct {
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
			errors.New("invalid azure.auth-mode. valid value is either aks, obo, or client-credential"),
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
	}
)

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
	}

	testData = append(testData, getTestDataForIndivitualError()...)

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			errs := test.opts.Validate()
			if test.expectedErr == nil {
				assert.Nil(t, errs)
			} else {
				if assert.NotNil(t, errs, "errors expected") {
					assert.EqualError(t, aggregator.NewAggregate(errs), aggregator.NewAggregate(test.expectedErr).Error())
				}
			}
		})
	}
}

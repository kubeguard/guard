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

package google

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
}{
	{
		"google.sa-json-file is empty",
		func(o Options) Options {
			o.ServiceAccountJsonFile = empty
			return o
		},
		errors.New("google.sa-json-file must be non-empty"),
	},
	{
		"google.admin-email is empty",
		func(o Options) Options {
			o.AdminEmail = empty
			return o
		},
		errors.New("google.admin-email must be non-empty"),
	},
}

func getNonEmptyOptions() Options {
	return Options{
		ServiceAccountJsonFile: nonempty,
		AdminEmail:             nonempty,
	}
}

func getEmptyOptions() Options {
	return Options{}
}

func getAllError() []error {
	var errs []error
	for _, d := range validationErrorData {
		errs = append(errs, d.expectedErr)
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
					assert.EqualError(t, utilerrors.NewAggregate(errs), utilerrors.NewAggregate(test.expectedErr).Error())
				}
			}
		})
	}
}

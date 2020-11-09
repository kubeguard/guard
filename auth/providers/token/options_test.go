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

package token

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

func TestOptionsValidate(t *testing.T) {
	validateData := struct {
		flagName string
		err      error
	}{
		flagName: "token-auth-file",
		err:      errors.New("token-auth-file must be non-empty"),
	}

	testdata := []struct {
		opts        Options
		expectedErr []error
	}{
		{Options{empty},
			[]error{validateData.err},
		},
		{
			Options{nonempty},
			nil,
		},
	}

	for _, test := range testdata {
		var testName string
		if test.opts.AuthFile == empty {
			testName = validateData.flagName + "empty"
		} else {
			testName = validateData.flagName + "non-empty"
		}

		t.Run(testName, func(t *testing.T) {
			errs := test.opts.Validate()
			if test.expectedErr == nil {
				assert.Nil(t, errs, "expected error nil")
			} else {
				if assert.NotNil(t, errs, "expected errors") {
					assert.EqualError(t, utilerrors.NewAggregate(errs), utilerrors.NewAggregate(test.expectedErr).Error(), "token auth options validation")
				}
			}
		})
	}
}

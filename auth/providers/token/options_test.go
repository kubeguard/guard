package token

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
					assert.EqualError(t, aggregator.NewAggregate(errs), aggregator.NewAggregate(test.expectedErr).Error(), "token auth options validation")
				}
			}
		})
	}
}

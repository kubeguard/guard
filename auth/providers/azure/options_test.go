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
	}{
		{
			"azure.client-secret must be non-empty",
			func(o Options) Options {
				o.ClientSecret = empty
				return o
			},
			errors.New("client secret must be non-empty"),
		},
		{
			"azure.client-id is empty",
			func(o Options) Options {
				o.ClientID = empty
				return o
			},
			errors.New("azure.client-id must be non-empty"),
		},
		{
			"azure.tenant-id is empty",
			func(o Options) Options {
				o.TenantID = empty
				return o
			},
			errors.New("azure.tenant-id must be non-empty"),
		},
	}
)

func getNonEmptyOptions() Options {
	return Options{
		ClientID:     nonempty,
		ClientSecret: nonempty,
		TenantID:     nonempty,
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
					assert.EqualError(t, aggregator.NewAggregate(errs), aggregator.NewAggregate(test.expectedErr).Error())
				}
			}
		})
	}
}

package ldap

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
			"ldap.server-address is empty",
			func(o Options) Options {
				o.ServerAddress = empty
				return o
			},
			errors.New("ldap.server-address must be non-empty"),
		},
		{
			"ldap.server-port is empty",
			func(o Options) Options {
				o.ServerPort = empty
				return o
			},
			errors.New("ldap.server-port must be non-empty"),
		},
		{
			"ldap.user-search-dn is empty",
			func(o Options) Options {
				o.UserSearchDN = empty
				return o
			},
			errors.New("ldap.user-search-dn must be non-empty"),
		},
		{
			"ldap.user-attribute is empty",
			func(o Options) Options {
				o.UserAttribute = empty
				return o
			},
			errors.New("ldap.user-attribute must be non-empty"),
		},
		{
			"ldap.group-search-dn is empty",
			func(o Options) Options {
				o.GroupSearchDN = empty
				return o
			},
			errors.New("ldap.group-search-dn must be non-empty"),
		},
		{
			"ldap.group-member-attribute is empty",
			func(o Options) Options {
				o.GroupMemberAttribute = empty
				return o
			},
			errors.New("ldap.group-member-attribute must be non-empty"),
		},
		{
			"ldap.group-name-attribute is empty",
			func(o Options) Options {
				o.GroupNameAttribute = empty
				return o
			},
			errors.New("ldap.group-name-attribute must be non-empty"),
		},
		{
			"ldap.is-secure-ldap and ldap.start-tls both are true",
			func(o Options) Options {
				o.IsSecureLDAP = true
				o.StartTLS = true
				return o
			},
			errors.New("ldap.is-secure-ldap and ldap.start-tls both can not be true at the same time"),
		},
		{
			"auth choice kerberos and ldap.keytab-file is empty",
			func(o Options) Options {
				o.KeytabFile = empty
				o.AuthenticationChoice = AuthChoiceKerberos
				return o
			},
			errors.New("for kerberos ldap.keytab-file must be non-empty"),
		},
	}
)

func getNonEmptyOptions() Options {
	return Options{
		ServerAddress:        nonempty,
		ServerPort:           nonempty,
		UserSearchDN:         nonempty,
		UserSearchFilter:     nonempty,
		UserAttribute:        nonempty,
		GroupSearchDN:        nonempty,
		GroupSearchFilter:    nonempty,
		GroupMemberAttribute: nonempty,
		GroupNameAttribute:   nonempty,
		IsSecureLDAP:         false,
		StartTLS:             false,
		AuthenticationChoice: AuthChoiceSimpleAuthentication,
		KeytabFile:           nonempty,
	}
}

func getEmptyOptions() Options {
	return Options{
		IsSecureLDAP:         true,
		StartTLS:             true,
		AuthenticationChoice: AuthChoiceKerberos,
	}
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

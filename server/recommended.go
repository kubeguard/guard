package server

import (
	"github.com/appscode/guard/auth/providers/azure"
	"github.com/appscode/guard/auth/providers/google"
	"github.com/appscode/guard/auth/providers/ldap"
	"github.com/appscode/guard/auth/providers/token"
	"github.com/spf13/pflag"
)

type RecommendedOptions struct {
	SecureServing SecureServingOptions
	NTP           NTPOptions
	Token         token.Options
	Google        google.Options
	Azure         azure.Options
	LDAP          ldap.Options
}

func NewRecommendedOptions() *RecommendedOptions {
	return &RecommendedOptions{
		SecureServing: NewSecureServingOptions(),
		NTP:           NewNTPOptions(),
	}
}

func (o *RecommendedOptions) AddFlags(fs *pflag.FlagSet) {
	o.SecureServing.AddFlags(fs)
	o.NTP.AddFlags(fs)
	o.Token.AddFlags(fs)
	o.Google.AddFlags(fs)
	o.Azure.AddFlags(fs)
	o.LDAP.AddFlags(fs)
}

func (o *RecommendedOptions) Validate() []error {
	var errors []error
	errors = append(errors, o.SecureServing.Validate()...)
	errors = append(errors, o.NTP.Validate()...)
	errors = append(errors, o.Token.Validate()...)
	errors = append(errors, o.Google.Validate()...)
	errors = append(errors, o.Azure.Validate()...)
	errors = append(errors, o.LDAP.Validate()...)

	return errors
}

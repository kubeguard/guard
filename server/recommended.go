package server

import (
	"github.com/appscode/guard/auth/providers"
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
	AuthProvider  providers.AuthProviders
}

func NewRecommendedOptions() *RecommendedOptions {
	return &RecommendedOptions{
		SecureServing: NewSecureServingOptions(),
		NTP:           NewNTPOptions(),
		Azure:         azure.NewOptions(),
		Token:         token.NewOptions(),
		Google:        google.NewOptions(),
		LDAP:          ldap.NewOptions(),
	}
}

func (o *RecommendedOptions) AddFlags(fs *pflag.FlagSet) {
	o.SecureServing.AddFlags(fs)
	o.NTP.AddFlags(fs)
	o.AuthProvider.AddFlags(fs)
	o.Token.AddFlags(fs)
	o.Google.AddFlags(fs)
	o.Azure.AddFlags(fs)
	o.LDAP.AddFlags(fs)
}

func (o *RecommendedOptions) Validate() []error {
	var errs []error
	errs = append(errs, o.SecureServing.Validate()...)
	errs = append(errs, o.NTP.Validate()...)
	errs = append(errs, o.AuthProvider.Validate()...)

	if o.AuthProvider.Has(token.OrgType) {
		errs = append(errs, o.Token.Validate()...)
	}
	if o.AuthProvider.Has(google.OrgType) {
		errs = append(errs, o.Google.Validate()...)
	}
	if o.AuthProvider.Has(azure.OrgType) {
		errs = append(errs, o.Azure.Validate()...)
	}
	if o.AuthProvider.Has(ldap.OrgType) {
		errs = append(errs, o.LDAP.Validate()...)
	}

	return errs
}

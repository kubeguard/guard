package server

import (
	"github.com/appscode/guard/auth/providers"
	"github.com/appscode/guard/auth/providers/azure"
	"github.com/appscode/guard/auth/providers/github"
	"github.com/appscode/guard/auth/providers/gitlab"
	"github.com/appscode/guard/auth/providers/google"
	"github.com/appscode/guard/auth/providers/ldap"
	"github.com/appscode/guard/auth/providers/token"
	"github.com/spf13/pflag"
)

type RecommendedOptions struct {
	SecureServing SecureServingOptions
	NTP           NTPOptions
	Github        github.Options
	Gitlab        gitlab.Options
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
		Github:        github.NewOptions(),
		Gitlab:        gitlab.NewOptions(),
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
	o.Github.AddFlags(fs)
	o.Gitlab.AddFlags(fs)
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

	if o.AuthProvider.Has(github.OrgType) {
		errs = append(errs, o.Github.Validate()...)
	}
	if o.AuthProvider.Has(gitlab.OrgType) {
		errs = append(errs, o.Gitlab.Validate()...)
	}
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

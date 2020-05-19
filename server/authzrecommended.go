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
package server

import (
	authz "github.com/appscode/guard/authz/providers"
	"github.com/appscode/guard/authz/providers/azure"

	"github.com/spf13/pflag"
)

type AuthzRecommendedOptions struct {
	Azure         azure.Options
	AuthzProvider authz.AuthzProviders
}

func NewAuthzRecommendedOptions() *AuthzRecommendedOptions {
	return &AuthzRecommendedOptions{
		Azure: azure.NewOptions(),
	}
}

func (o *AuthzRecommendedOptions) AddFlags(fs *pflag.FlagSet) {
	o.Azure.AddFlags(fs)
	o.AuthzProvider.AddFlags(fs)
}

func (o *AuthzRecommendedOptions) Validate(opts *AuthRecommendedOptions) []error {
	var errs []error
	if len(o.AuthzProvider.Providers) > 0 {
		errs = append(errs, o.AuthzProvider.Validate()...)
	}

	if o.AuthzProvider.Has(azure.OrgType) {
		errs = append(errs, o.Azure.Validate(opts.Azure)...)
	}

	return errs
}

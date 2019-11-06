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
package providers

import (
	"fmt"
	"strings"

	"github.com/appscode/guard/auth"
	_ "github.com/appscode/guard/auth/providers/azure"
	_ "github.com/appscode/guard/auth/providers/github"
	_ "github.com/appscode/guard/auth/providers/gitlab"
	_ "github.com/appscode/guard/auth/providers/google"
	_ "github.com/appscode/guard/auth/providers/ldap"
	_ "github.com/appscode/guard/auth/providers/token"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AuthProviders struct {
	Providers []string // contains providers name for which guard will provide service, required
}

func (a *AuthProviders) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&a.Providers, "auth-providers", a.Providers, fmt.Sprintf("name of providers for which guard will provide authentication service (required), supported providers : %v", auth.SupportedOrgs.String()))
}

func (a *AuthProviders) Validate() []error {
	var errs []error
	if len(a.Providers) == 0 {
		errs = append(errs, errors.New("auth-providers must be non-empty"))
	}

	for _, p := range a.Providers {
		if !auth.SupportedOrgs.Has(p) {
			errs = append(errs, errors.Errorf("provider %s not supported", p))
		}
	}
	return errs
}

func (a *AuthProviders) Apply(d *apps.Deployment) (extraObjs []runtime.Object, err error) {
	if len(a.Providers) > 0 {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--auth-providers=%s", strings.Join(a.Providers, ",")))
	}

	return nil, nil
}

func (a *AuthProviders) Has(name string) bool {
	name = strings.TrimSpace(name)
	for _, p := range a.Providers {
		if strings.EqualFold(p, name) {
			return true
		}
	}
	return false
}

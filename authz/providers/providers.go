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

	"go.kubeguard.dev/guard/authz"
	_ "go.kubeguard.dev/guard/authz/providers/azure"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AuthzProviders struct {
	Providers []string // contains providers name for which guard will provide service, required
}

func (a *AuthzProviders) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&a.Providers, "authz-providers", a.Providers, fmt.Sprintf("name of providers for which guard will provide authorization service, supported providers : %v", authz.SupportedOrgs.String()))
}

func (a *AuthzProviders) Validate() []error {
	var errs []error

	for _, p := range a.Providers {
		if !authz.SupportedOrgs.Has(p) {
			errs = append(errs, errors.Errorf("provider %s not supported", p))
		}
	}
	return errs
}

func (a *AuthzProviders) Apply(d *apps.Deployment) (extraObjs []runtime.Object, err error) {
	if len(a.Providers) > 0 {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--authz-providers=%s", strings.Join(a.Providers, ",")))
	}

	return nil, nil
}

func (a *AuthzProviders) Has(name string) bool {
	name = strings.TrimSpace(name)
	for _, p := range a.Providers {
		if strings.EqualFold(p, name) {
			return true
		}
	}
	return false
}

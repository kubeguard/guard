package providers

import (
	"fmt"
	"strings"

	"github.com/appscode/guard/auth"
	_ "github.com/appscode/guard/auth/providers/appscode"
	_ "github.com/appscode/guard/auth/providers/azure"
	_ "github.com/appscode/guard/auth/providers/github"
	_ "github.com/appscode/guard/auth/providers/gitlab"
	_ "github.com/appscode/guard/auth/providers/google"
	_ "github.com/appscode/guard/auth/providers/ldap"
	_ "github.com/appscode/guard/auth/providers/token"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/api/apps/v1beta1"
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

func (a *AuthProviders) Apply(d *v1beta1.Deployment) (extraObjs []runtime.Object, err error) {
	if len(a.Providers) > 0 {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--auth-providers=%s", strings.Join(a.Providers, ",")))
	}

	return nil, nil
}

func (a *AuthProviders) Has(name string) bool {
	name = strings.TrimSpace(strings.ToLower(name))
	for _, p := range a.Providers {
		if strings.ToLower(p) == name {
			return true
		}
	}
	return false
}

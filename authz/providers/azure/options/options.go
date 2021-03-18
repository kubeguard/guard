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

package options

import (
	"fmt"
	"strings"

	"github.com/appscode/guard/auth/providers/azure"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	AKSAuthzMode               = "aks"
	ARCAuthzMode               = "arc"
	defaultArmCallLimit        = 2000
	maxPermissibleArmCallLimit = 4000
)

type Options struct {
	AuthzMode                      string
	ResourceId                     string
	AKSAuthzTokenURL               string
	ARMCallLimit                   int
	SkipAuthzCheckConfig           string
	SkipAuthzForNonAADUsers        bool
	AllowNonResDiscoveryPathAccess bool
}

func NewOptions() Options {
	return Options{
		ARMCallLimit:                   defaultArmCallLimit,
		SkipAuthzCheckConfig:           "",
		SkipAuthzForNonAADUsers:        true,
		AllowNonResDiscoveryPathAccess: true}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.AuthzMode, "azure.authz-mode", "", "authz mode to call RBAC api, valid value is either aks or arc")
	fs.StringVar(&o.ResourceId, "azure.resource-id", "", "azure cluster resource id (//subscription/<subName>/resourcegroups/<RGname>/providers/Microsoft.ContainerService/managedClusters/<clustername> for AKS or //subscription/<subName>/resourcegroups/<RGname>/providers/Microsoft.Kubernetes/connectedClusters/<clustername> for arc) to be used as scope for RBAC check")
	fs.StringVar(&o.AKSAuthzTokenURL, "azure.aks-authz-token-url", "", "url to call for AKS Authz flow")
	fs.IntVar(&o.ARMCallLimit, "azure.arm-call-limit", o.ARMCallLimit, "No of calls before which webhook switch to new ARM instance to avoid throttling")
	fs.StringVar(&o.SkipAuthzCheckConfig, "azure.skip-authz-check-config", o.SkipAuthzCheckConfig, "path of config file which contains names of usernames/email. Azure Authz check will be skipped for these users")
	fs.BoolVar(&o.SkipAuthzForNonAADUsers, "azure.skip-authz-for-non-aad-users", o.SkipAuthzForNonAADUsers, "skip authz for non AAD users")
	fs.BoolVar(&o.AllowNonResDiscoveryPathAccess, "azure.allow-nonres-discovery-path-access", o.AllowNonResDiscoveryPathAccess, "allow access on Non Resource paths required for discovery, setting it false will require explicit non resource path role assignment for all users in Azure RBAC")
}

func (o *Options) Validate(azure azure.Options) []error {
	var errs []error
	o.AuthzMode = strings.ToLower(o.AuthzMode)
	switch o.AuthzMode {
	case AKSAuthzMode:
	case ARCAuthzMode:
	default:
		errs = append(errs, errors.New("invalid azure.authz-mode. valid value is either aks or arc"))
	}

	if o.AuthzMode != "" && o.ResourceId == "" {
		errs = append(errs, errors.New("azure.resource-id must be non-empty for authorization"))
	}

	if o.AuthzMode == AKSAuthzMode && o.AKSAuthzTokenURL == "" {
		errs = append(errs, errors.New("azure.aks-authz-token-url must be non-empty"))
	}

	if o.AuthzMode != AKSAuthzMode && o.AKSAuthzTokenURL != "" {
		errs = append(errs, errors.New("azure.aks-authz-token-url must be set only with AKS authz mode"))
	}

	if o.AuthzMode == ARCAuthzMode {
		if azure.ClientSecret == "" {
			errs = append(errs, errors.New("azure.client-secret must be non-empty"))
		}
		if azure.ClientID == "" {
			errs = append(errs, errors.New("azure.client-id must be non-empty"))
		}
	}

	if o.ARMCallLimit > maxPermissibleArmCallLimit {
		errs = append(errs, fmt.Errorf("azure.arm-call-limit must not be more than %d", maxPermissibleArmCallLimit))
	}

	return errs
}

func (o Options) Apply(d *apps.Deployment) (extraObjs []runtime.Object, err error) {
	args := d.Spec.Template.Spec.Containers[0].Args
	switch o.AuthzMode {
	case AKSAuthzMode:
		fallthrough
	case ARCAuthzMode:
		args = append(args, fmt.Sprintf("--azure.authz-mode=%s", o.AuthzMode))
		args = append(args, fmt.Sprintf("--azure.resource-id=%s", o.ResourceId))
		args = append(args, fmt.Sprintf("--azure.arm-call-limit=%d", o.ARMCallLimit))
	}

	if o.AKSAuthzTokenURL != "" {
		args = append(args, fmt.Sprintf("--azure.aks-authz-token-url=%s", o.AKSAuthzTokenURL))
	}

	if o.SkipAuthzCheckConfig != "" {
		args = append(args, fmt.Sprintf("--azure.skip-authz-check-config=%s", o.SkipAuthzCheckConfig))
	}

	args = append(args, fmt.Sprintf("--azure.skip-authz-for-non-aad-users=%t", o.SkipAuthzForNonAADUsers))

	args = append(args, fmt.Sprintf("--azure.allow-nonres-discovery-path-access=%t", o.AllowNonResDiscoveryPathAccess))

	d.Spec.Template.Spec.Containers[0].Args = args
	return extraObjs, nil
}

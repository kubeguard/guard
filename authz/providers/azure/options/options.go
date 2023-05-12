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
	"time"

	"go.kubeguard.dev/guard/auth/providers/azure"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	AKSAuthzMode               = "aks"
	ARCAuthzMode               = "arc"
	FleetAuthzMode             = "fleet"
	defaultArmCallLimit        = 2000
	maxPermissibleArmCallLimit = 4000
)

type Options struct {
	AuthzMode                           string
	ResourceId                          string
	AKSAuthzTokenURL                    string
	ARMCallLimit                        int
	SkipAuthzCheck                      []string
	SkipAuthzForNonAADUsers             bool
	AllowNonResDiscoveryPathAccess      bool
	UseNamespaceResourceScopeFormat     bool
	DiscoverResources                   bool
	ReconcileDiscoverResourcesFrequency time.Duration
	KubeConfigFile                      string
}

func NewOptions() Options {
	return Options{
		ARMCallLimit:                        defaultArmCallLimit,
		SkipAuthzCheck:                      []string{""},
		SkipAuthzForNonAADUsers:             true,
		AllowNonResDiscoveryPathAccess:      true,
		UseNamespaceResourceScopeFormat:     false,
		DiscoverResources:                   false,
		ReconcileDiscoverResourcesFrequency: 5 * time.Minute,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.AuthzMode, "azure.authz-mode", "", "authz mode to call RBAC api, valid values are either aks, arc, or fleet")
	fs.StringVar(&o.ResourceId, "azure.resource-id", "", "azure cluster resource id (//subscription/<subName>/resourcegroups/<RGname>/providers/Microsoft.ContainerService/managedClusters/<clustername> for AKS, //subscription/<subName>/resourcegroups/<RGname>/providers/Microsoft.ContainerService/fleets/<clustername> for Azure Kubernetes Fleet Manager, or //subscription/<subName>/resourcegroups/<RGname>/providers/Microsoft.Kubernetes/connectedClusters/<clustername> for arc) to be used as scope for RBAC check")
	fs.StringVar(&o.AKSAuthzTokenURL, "azure.aks-authz-token-url", "", "url to call for AKS Authz flow")
	fs.IntVar(&o.ARMCallLimit, "azure.arm-call-limit", o.ARMCallLimit, "No of calls before which webhook switch to new ARM instance to avoid throttling")
	fs.StringSliceVar(&o.SkipAuthzCheck, "azure.skip-authz-check", o.SkipAuthzCheck, "name of usernames/email for which authz check will be skipped")
	fs.BoolVar(&o.SkipAuthzForNonAADUsers, "azure.skip-authz-for-non-aad-users", o.SkipAuthzForNonAADUsers, "skip authz for non AAD users")
	fs.BoolVar(&o.AllowNonResDiscoveryPathAccess, "azure.allow-nonres-discovery-path-access", o.AllowNonResDiscoveryPathAccess, "allow access on Non Resource paths required for discovery, setting it false will require explicit non resource path role assignment for all users in Azure RBAC")
	fs.BoolVar(&o.UseNamespaceResourceScopeFormat, "azure.use-ns-resource-scope-format", o.UseNamespaceResourceScopeFormat, "use namespace as resource scope format for making rbac checkaccess calls at namespace scope")
	fs.StringVar(&o.KubeConfigFile, "azure.kubeconfig-file", "", "path to the kubeconfig of cluster.")
	fs.BoolVar(&o.DiscoverResources, "azure.discover-resources", o.DiscoverResources, "fetch list of resources and operations from apiserver and azure. Default: false")
	fs.DurationVar(&o.ReconcileDiscoverResourcesFrequency, "azure.discover-resources-frequency", o.ReconcileDiscoverResourcesFrequency, "Frequency at which discover resources should be reconciled. Default: 5m")
}

func (o *Options) Validate(azure azure.Options) []error {
	var errs []error
	o.AuthzMode = strings.ToLower(o.AuthzMode)
	switch o.AuthzMode {
	case AKSAuthzMode:
	case ARCAuthzMode:
	case FleetAuthzMode:
	default:
		errs = append(errs, errors.New("invalid azure.authz-mode. valid value is either aks, arc, or fleet"))
	}

	if o.AuthzMode != "" && o.ResourceId == "" {
		errs = append(errs, errors.New("azure.resource-id must be non-empty for authorization"))
	}

	if (o.AuthzMode == AKSAuthzMode || o.AuthzMode == FleetAuthzMode) && o.AKSAuthzTokenURL == "" {
		errs = append(errs, errors.New("azure.aks-authz-token-url must be non-empty"))
	}

	if o.AuthzMode != AKSAuthzMode && o.AuthzMode != FleetAuthzMode && o.AKSAuthzTokenURL != "" {
		errs = append(errs, errors.New("azure.aks-authz-token-url must be set only with AKS/Fleet authz mode"))
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

	if len(o.SkipAuthzCheck) > 0 {
		args = append(args, fmt.Sprintf("--azure.skip-authz-check=%s", strings.Join(o.SkipAuthzCheck, ",")))
	}

	args = append(args, fmt.Sprintf("--azure.skip-authz-for-non-aad-users=%t", o.SkipAuthzForNonAADUsers))

	args = append(args, fmt.Sprintf("--azure.allow-nonres-discovery-path-access=%t", o.AllowNonResDiscoveryPathAccess))

	d.Spec.Template.Spec.Containers[0].Args = args
	return extraObjs, nil
}

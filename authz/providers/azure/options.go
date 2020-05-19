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

package azure

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
	AuthzMode                    string
	ResourceId                   string
	AKSAuthzURL                  string
	ARMCallLimit                 int
	SkipAuthzCheck               []string
	AuthzResolveGroupMemberships bool
	SkipAuthzForNonAADUsers      bool
}

func NewOptions() Options {
	return Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.AuthzMode, "azure.authz-mode", "", "authz mode to call RBAC api, valid value is either aks or arc")
	fs.StringVar(&o.ResourceId, "azure.resource-id", "", "azure cluster resource id (//subscription/<subName>/resourcegroups/<RGname>/providers/Microsoft.ContainerService/managedClusters/<clustername> for AKS or //subscription/<subName>/resourcegroups/<RGname>/providers/Microsoft.Kubernetes/connectedClusters/<clustername> for arc) to be used as scope for RBAC check")
	fs.StringVar(&o.AKSAuthzURL, "azure.aks-authz-url", "", "url to call for AKS Authz flow")
	fs.IntVar(&o.ARMCallLimit, "azure.arm-call-limit", defaultArmCallLimit, "No of calls before which webhook switch to new ARM instance to avoid throttling")
	fs.StringSliceVar(&o.SkipAuthzCheck, "azure.skip-authz-check", []string{""}, "name of usernames/email for which authz check will be skipped")
	fs.BoolVar(&o.AuthzResolveGroupMemberships, "azure.authz-resolve-group-memberships", o.AuthzResolveGroupMemberships, "set to true to resolve group membership by authorizer. Setting to false will use group list from subjectaccessreview request")
	fs.BoolVar(&o.SkipAuthzForNonAADUsers, "azure.skip-authz-for-non-aad-users", o.SkipAuthzForNonAADUsers, "skip authz for non AAD users")
}

func (o *Options) Validate(azure azure.Options) []error {
	var errs []error
	o.AuthzMode = strings.ToLower(o.AuthzMode)
	switch o.AuthzMode {
	case AKSAuthzMode:
	case ARCAuthzMode:
	case "":
	default:
		errs = append(errs, errors.New("invalid azure.authz-mode. valid value is either aks or arc"))
	}

	if o.AuthzMode != "" && o.ResourceId == "" {
		errs = append(errs, errors.New("azure.resource-id must be non-empty for authorization"))
	}

	if o.AuthzMode == AKSAuthzMode && o.AKSAuthzURL == "" {
		errs = append(errs, errors.New("azure.aks-authz-url must be non-empty"))
	}

	if o.AuthzMode != AKSAuthzMode && o.AKSAuthzURL != "" {
		errs = append(errs, errors.New("azure.aks-authz-url must be set only with AKS authz mode"))
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
		errs = append(errs, errors.New("azure.arm-call-limit must not be more than 4000"))
	}

	return errs
}

func (o Options) Apply(d *apps.Deployment) (extraObjs []runtime.Object, err error) {
	container := d.Spec.Template.Spec.Containers[0]
	args := container.Args
	switch o.AuthzMode {
	case AKSAuthzMode:
		fallthrough
	case ARCAuthzMode:
		args = append(args, fmt.Sprintf("--azure.authz-mode=%s", o.AuthzMode))
		args = append(args, fmt.Sprintf("--azure.resource-id=%s", o.ResourceId))
		args = append(args, fmt.Sprintf("--azure.arm-call-limit=%d", o.ARMCallLimit))
	}

	if o.AKSAuthzURL != "" {
		args = append(args, fmt.Sprintf("--azure.aks-authz-url=%s", o.AKSAuthzURL))
	}

	if len(o.SkipAuthzCheck) > 0 {
		args = append(args, fmt.Sprintf("--azure.skip-authz-check=%s", strings.Join(o.SkipAuthzCheck, ",")))
	}

	args = append(args, fmt.Sprintf("--azure.authz-resolve-group-memberships=%t", o.AuthzResolveGroupMemberships))

	args = append(args, fmt.Sprintf("--azure.skip-authz-for-non-aad-users=%t", o.SkipAuthzForNonAADUsers))

	container.Args = args
	d.Spec.Template.Spec.Containers[0] = container
	return extraObjs, nil
}

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
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	AKSAuthMode              = "aks"
	OBOAuthMode              = "obo"
	ClientCredentialAuthMode = "client-credential"
	PassthroughAuthMode      = "passthrough"
)

type Options struct {
	Environment                              string
	ClientID                                 string
	ClientSecret                             string
	TenantID                                 string
	UseGroupUID                              bool
	AuthMode                                 string
	AKSTokenURL                              string
	EnablePOP                                bool
	POPTokenHostname                         string
	POPTokenValidTill                        time.Duration
	ResolveGroupMembershipOnlyOnOverageClaim bool
	VerifyClientID                           bool
}

func NewOptions() Options {
	return Options{
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		UseGroupUID:  true,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Environment, "azure.environment", o.Environment, "Azure cloud environment")
	fs.StringVar(&o.ClientID, "azure.client-id", o.ClientID, "MS Graph application client ID to use")
	fs.StringVar(&o.ClientSecret, "azure.client-secret", o.ClientSecret, "MS Graph application client secret to use")
	fs.StringVar(&o.TenantID, "azure.tenant-id", o.TenantID, "MS Graph application tenant id to use")
	fs.BoolVar(&o.UseGroupUID, "azure.use-group-uid", o.UseGroupUID, "Use group UID for authentication instead of group display name")
	fs.StringVar(&o.AuthMode, "azure.auth-mode", "client-credential", "auth mode to call graph api, valid value is either aks, obo, or client-credential")
	fs.StringVar(&o.AKSTokenURL, "azure.aks-token-url", "", "url to call for AKS OBO flow")
	fs.StringVar(&o.POPTokenHostname, "azure.pop-hostname", "", "hostname used to run the pop hostname verification; 'u' claim")
	fs.BoolVar(&o.EnablePOP, "azure.enable-pop", false, "Enabling pop token verification")
	fs.DurationVar(&o.POPTokenValidTill, "azure.pop-token-validtill", 15, "time duration till what PoP token considered valid from creation time, default 15 min")
	fs.BoolVar(&o.ResolveGroupMembershipOnlyOnOverageClaim, "azure.graph-call-on-overage-claim", o.ResolveGroupMembershipOnlyOnOverageClaim, "set to true to resolve group membership only when overage claim is present. setting to false will always call graph api to resolve group membership")
	fs.BoolVar(&o.VerifyClientID, "azure.verify-clientID", o.VerifyClientID, "set to true to validate token's audience claim matches clientID")
}

func (o *Options) Validate() []error {
	var errs []error
	o.AuthMode = strings.ToLower(o.AuthMode)
	switch o.AuthMode {
	case AKSAuthMode:
	case OBOAuthMode:
	case ClientCredentialAuthMode:
	case PassthroughAuthMode:
	default:
		errs = append(errs, errors.New("invalid azure.auth-mode. valid value is either aks, obo, client-credential or pop"))
	}

	if o.AuthMode != AKSAuthMode && o.AuthMode != PassthroughAuthMode {
		if o.ClientSecret == "" {
			errs = append(errs, errors.New("azure.client-secret must be non-empty"))
		}
	}
	if o.AuthMode == AKSAuthMode && o.AKSTokenURL == "" {
		errs = append(errs, errors.New("azure.aks-token-url must be non-empty"))
	}
	if o.TenantID == "" {
		errs = append(errs, errors.New("azure.tenant-id must be non-empty"))
	}
	if o.VerifyClientID && o.ClientID == "" {
		errs = append(errs, errors.New("azure.client-id must be non-empty when azure.verify-clientID is set"))
	}
	if o.EnablePOP {
		if o.POPTokenHostname == "" {
			errs = append(errs, errors.New("azure.pop-hostname must be non-empty"))
		}
	}
	return errs
}

func (o Options) Apply(d *apps.Deployment) (extraObjs []runtime.Object, err error) {
	container := d.Spec.Template.Spec.Containers[0]

	// create auth secret
	authSecret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-azure-auth",
			Namespace: d.Namespace,
			Labels:    d.Labels,
		},
		Data: map[string][]byte{
			"client-secret": []byte(o.ClientSecret),
		},
	}
	extraObjs = append(extraObjs, authSecret)

	// mount auth secret into deployment
	volMount := core.VolumeMount{
		Name:      authSecret.Name,
		MountPath: "/etc/guard/auth/azure",
	}
	container.VolumeMounts = append(container.VolumeMounts, volMount)

	vol := core.Volume{
		Name: authSecret.Name,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName:  authSecret.Name,
				DefaultMode: pointer.Int32P(0555),
			},
		},
	}
	d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, vol)

	// use auth secret in container[0] args
	container.Env = append(container.Env, core.EnvVar{
		Name: "AZURE_CLIENT_SECRET",
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: authSecret.Name,
				},
				Key: "client-secret",
			},
		},
	})

	args := container.Args
	if o.Environment != "" {
		args = append(args, fmt.Sprintf("--azure.environment=%s", o.Environment))
	}
	if o.ClientID != "" {
		args = append(args, fmt.Sprintf("--azure.client-id=%s", o.ClientID))
	}
	if o.TenantID != "" {
		args = append(args, fmt.Sprintf("--azure.tenant-id=%s", o.TenantID))
	}

	switch o.AuthMode {
	case AKSAuthMode:
		fallthrough
	case OBOAuthMode:
		fallthrough
	case ClientCredentialAuthMode:
		args = append(args, fmt.Sprintf("--azure.auth-mode=%s", o.AuthMode))
	default:
		args = append(args, fmt.Sprintf("--azure.auth-mode=%s", ClientCredentialAuthMode))
	}

	if o.AKSTokenURL != "" {
		args = append(args, fmt.Sprintf("--azure.aks-token-url=%s", o.AKSTokenURL))
	}

	args = append(args, fmt.Sprintf("--azure.use-group-uid=%t", o.UseGroupUID))

	args = append(args, fmt.Sprintf("--azure.graph-call-on-overage-claim=%t", o.ResolveGroupMembershipOnlyOnOverageClaim))

	args = append(args, fmt.Sprintf("--azure.verify-clientID=%t", o.VerifyClientID))

	container.Args = args
	d.Spec.Template.Spec.Containers[0] = container

	return extraObjs, nil
}

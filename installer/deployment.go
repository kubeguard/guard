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

package installer

import (
	"fmt"
	"strings"

	"go.kubeguard.dev/guard/auth/providers/azure"
	"go.kubeguard.dev/guard/auth/providers/github"
	"go.kubeguard.dev/guard/auth/providers/gitlab"
	"go.kubeguard.dev/guard/auth/providers/google"
	"go.kubeguard.dev/guard/auth/providers/ldap"
	"go.kubeguard.dev/guard/auth/providers/token"
	azureauthz "go.kubeguard.dev/guard/authz/providers/azure"
	"go.kubeguard.dev/guard/server"

	"gomodules.xyz/pointer"
	stringz "gomodules.xyz/x/strings"
	v "gomodules.xyz/x/version"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	DefaultAzureEntraSDKImage  = "mcr.microsoft.com/entra-sdk/auth-sidecar:1.0.0-azurelinux3.0-distroless"
	azureEntraSDKContainerName = "entra-sdk"
	azureEntraSDKPort          = 8080
	azureLinuxBaseCoreImage    = "mcr.microsoft.com/azurelinux/base/core:3.0"
	proxyCertStoreVolumeName   = "proxy-certstore"
	guardSSLCertsVolumeName    = "ssl-certs"
	entraSDKCertsVolumeName    = "entra-sdk-ssl-certs"
)

func newDeployment(authopts AuthOptions, authzopts AuthzOptions) (objects []runtime.Object, err error) {
	d := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: authopts.Namespace,
			Labels:    labels,
		},
		Spec: apps.DeploymentSpec{
			Replicas: pointer.Int32P(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: core.PodSpec{
					PriorityClassName: "system-cluster-critical",
					Containers: []core.Container{
						{
							Name:  "guard",
							Image: authopts.guardImage(),
							Args: []string{
								"run",
								fmt.Sprintf("--v=%s", stringz.Val(authopts.VerbosityLevel, "3")),
							},
							Ports: []core.ContainerPort{
								{
									ContainerPort: server.ServingPort,
								},
							},
							ReadinessProbe: &core.Probe{
								ProbeHandler: core.ProbeHandler{
									HTTPGet: &core.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(server.ServingPort),
										Scheme: core.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: 1,
							},
							LivenessProbe: &core.Probe{
								ProbeHandler: core.ProbeHandler{
									HTTPGet: &core.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(server.ServingPort),
										Scheme: core.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: 1,
							},
						},
					},
					Tolerations: []core.Toleration{
						{
							Key:      "CriticalAddonsOnly",
							Operator: core.TolerationOpExists,
						},
					},
				},
			},
		},
	}
	if authopts.imagePullSecret != "" {
		d.Spec.Template.Spec.ImagePullSecrets = []core.LocalObjectReference{
			{
				Name: authopts.imagePullSecret,
			},
		}
	}
	if authopts.isLocalGuardImage() {
		d.Spec.Template.Spec.Containers[0].ImagePullPolicy = core.PullNever
	}
	if authopts.RunOnMaster {
		d.Spec.Template.Spec.NodeSelector = map[string]string{
			"node-role.kubernetes.io/master": "",
		}
		d.Spec.Template.Spec.Tolerations = append(d.Spec.Template.Spec.Tolerations, core.Toleration{
			Key:      "node-role.kubernetes.io/master",
			Operator: core.TolerationOpExists,
			Effect:   core.TaintEffectNoSchedule,
		})
	}
	if authopts.AuthProvider.Has(azure.OrgType) && authopts.UseAzureEntraSDK {
		entraSDKEnv, err := authopts.Azure.EntraSDKEnvVars()
		if err != nil {
			return nil, err
		}
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, core.Container{
			Name:  azureEntraSDKContainerName,
			Image: authopts.AzureEntraSDKImage,
			Env:   entraSDKEnv,
			ReadinessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					HTTPGet: &core.HTTPGetAction{
						Path:   "/healthz",
						Port:   intstr.FromInt(azureEntraSDKPort),
						Scheme: core.URISchemeHTTP,
						HTTPHeaders: []core.HTTPHeader{{
							Name:  "Host",
							Value: "localhost",
						}},
					},
				},
				InitialDelaySeconds: 1,
			},
			LivenessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					HTTPGet: &core.HTTPGetAction{
						Path:   "/healthz",
						Port:   intstr.FromInt(azureEntraSDKPort),
						Scheme: core.URISchemeHTTP,
						HTTPHeaders: []core.HTTPHeader{{
							Name:  "Host",
							Value: "localhost",
						}},
					},
				},
				InitialDelaySeconds: 1,
			},
		})
	}
	if authopts.HttpsProxy != "" || authopts.HttpProxy != "" || authopts.NoProxy != "" {
		proxyEnvVarRef := core.EnvFromSource{
			SecretRef: &core.SecretEnvSource{},
		}

		proxyEnvVarRef.SecretRef.Name = "guard-proxy"
		d.Spec.Template.Spec.Containers[0].EnvFrom = append(d.Spec.Template.Spec.Containers[0].EnvFrom, proxyEnvVarRef)
		if authopts.UseAzureEntraSDK {
			d.Spec.Template.Spec.Containers[1].EnvFrom = append(d.Spec.Template.Spec.Containers[1].EnvFrom, proxyEnvVarRef)
		}
		if authopts.ProxyCert != "" {
			proxycertVolumeRef := core.Volume{
				Name: proxyCertStoreVolumeName,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: "guard-proxy-cert",
					},
				},
			}

			sslCertsVolumeRef := core.Volume{
				Name: guardSSLCertsVolumeName,
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			}

			d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, proxycertVolumeRef)
			d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, sslCertsVolumeRef)

			sslCertsVolumeMount := core.VolumeMount{
				Name:      guardSSLCertsVolumeName,
				MountPath: "/etc/ssl/certs/",
			}

			d.Spec.Template.Spec.Containers[0].VolumeMounts = append(d.Spec.Template.Spec.Containers[0].VolumeMounts, sslCertsVolumeMount)

			optionalSecret := false
			initContainers := []core.Container{
				{
					Name:  "update-proxy-certs",
					Image: "nginx:stable-alpine",
					Command: []string{
						"sh",
						"-c",
						"update-ca-certificates",
					},
					EnvFrom: []core.EnvFromSource{
						{
							SecretRef: &core.SecretEnvSource{
								LocalObjectReference: core.LocalObjectReference{
									Name: "guard-proxy",
								},
								Optional: &optionalSecret,
							},
						},
					},
					VolumeMounts: []core.VolumeMount{
						{
							Name:      proxyCertStoreVolumeName,
							MountPath: "/usr/local/share/ca-certificates/proxy-cert.crt",
							SubPath:   "proxy-cert.crt",
						},
						{
							Name:      guardSSLCertsVolumeName,
							MountPath: "/etc/ssl/certs/",
						},
					},
				},
			}
			if authopts.UseAzureEntraSDK {
				d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, core.Volume{
					Name: entraSDKCertsVolumeName,
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				})

				d.Spec.Template.Spec.Containers[1].VolumeMounts = append(d.Spec.Template.Spec.Containers[1].VolumeMounts, core.VolumeMount{
					Name:      entraSDKCertsVolumeName,
					MountPath: "/etc/pki/ca-trust/extracted/",
				})

				initContainers = append(initContainers, core.Container{
					Name:  "update-entra-sdk-proxy-certs",
					Image: azureLinuxBaseCoreImage,
					Command: []string{
						"sh",
						"-c",
						"mkdir -p /etc/pki/ca-trust/extracted/openssl /etc/pki/ca-trust/extracted/pem /etc/pki/ca-trust/extracted/java /etc/pki/ca-trust/extracted/edk2 && update-ca-trust",
					},
					EnvFrom: []core.EnvFromSource{
						{
							SecretRef: &core.SecretEnvSource{
								LocalObjectReference: core.LocalObjectReference{
									Name: "guard-proxy",
								},
								Optional: &optionalSecret,
							},
						},
					},
					VolumeMounts: []core.VolumeMount{
						{
							Name:      proxyCertStoreVolumeName,
							MountPath: "/etc/pki/ca-trust/source/anchors/proxy-cert.crt",
							SubPath:   "proxy-cert.crt",
						},
						{
							Name:      entraSDKCertsVolumeName,
							MountPath: "/etc/pki/ca-trust/extracted/",
						},
					},
				})
			}

			d.Spec.Template.Spec.InitContainers = initContainers
		}
	}
	objects = append(objects, d)

	servingOpts := server.NewSecureServingOptionsFromDir(authopts.PkiDir)
	if extras, err := servingOpts.Apply(d); err != nil {
		return nil, err
	} else {
		objects = append(objects, extras...)
	}

	if extras, err := authopts.AuthProvider.Apply(d); err != nil {
		return nil, err
	} else {
		objects = append(objects, extras...)
	}

	if authopts.AuthProvider.Has(token.OrgType) {
		if extras, err := authopts.Token.Apply(d); err != nil {
			return nil, err
		} else {
			objects = append(objects, extras...)
		}
	}

	if authopts.AuthProvider.Has(google.OrgType) {
		if extras, err := authopts.Google.Apply(d); err != nil {
			return nil, err
		} else {
			objects = append(objects, extras...)
		}
	}

	if authopts.AuthProvider.Has(azure.OrgType) {
		azureOpts := authopts.Azure
		if authopts.UseAzureEntraSDK {
			azureOpts.EntraSDKURL = fmt.Sprintf("http://127.0.0.1:%d", azureEntraSDKPort)
		}
		if extras, err := azureOpts.Apply(d); err != nil {
			return nil, err
		} else {
			objects = append(objects, extras...)
		}
	}

	if authopts.AuthProvider.Has(ldap.OrgType) {
		if extras, err := authopts.LDAP.Apply(d); err != nil {
			return nil, err
		} else {
			objects = append(objects, extras...)
		}
	}

	if authopts.AuthProvider.Has(github.OrgType) {
		if extras, err := authopts.Github.Apply(d); err != nil {
			return nil, err
		} else {
			objects = append(objects, extras...)
		}
	}

	if authopts.AuthProvider.Has(gitlab.OrgType) {
		if extras, err := authopts.Gitlab.Apply(d); err != nil {
			return nil, err
		} else {
			objects = append(objects, extras...)
		}
	}

	if len(authzopts.AuthzProvider.Providers) > 0 {
		if extras, err := authzopts.AuthzProvider.Apply(d); err != nil {
			return nil, err
		} else {
			objects = append(objects, extras...)
		}
	}

	if authzopts.AuthzProvider.Has(azureauthz.OrgType) {
		if extras, err := authzopts.Azure.Apply(d); err != nil {
			return nil, err
		} else {
			objects = append(objects, extras...)
		}
	}

	return
}

func (o AuthOptions) guardImage() string {
	if o.GuardImage != "" {
		return o.GuardImage
	}

	return fmt.Sprintf("%s/guard:%v", o.PrivateRegistry, stringz.Val(v.Version.Version, "canary"))
}

func (o AuthOptions) isLocalGuardImage() bool {
	if o.GuardImage == "" {
		return false
	}

	registry := o.GuardImage
	if slash := strings.Index(registry, "/"); slash >= 0 {
		registry = registry[:slash]
	}

	return registry == "localhost" || strings.HasPrefix(registry, "localhost:") || registry == "127.0.0.1" ||
		strings.HasPrefix(registry, "127.0.0.1:") || registry == "[::1]" || strings.HasPrefix(registry, "[::1]:")
}

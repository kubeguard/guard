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
							Image: fmt.Sprintf("%s/guard:%v", authopts.PrivateRegistry, stringz.Val(v.Version.Version, "canary")),
							Args: []string{
								"run",
								fmt.Sprintf("--v=%s", authopts.VerbosityLevel),
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
								InitialDelaySeconds: int32(30),
							},
							LivenessProbe: &core.Probe{
								ProbeHandler: core.ProbeHandler{
									HTTPGet: &core.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(server.ServingPort),
										Scheme: core.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: int32(30),
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
	if authopts.HttpsProxy != "" || authopts.HttpProxy != "" || authopts.NoProxy != "" {
		proxyEnvVarRef := core.EnvFromSource{
			SecretRef: &core.SecretEnvSource{},
		}

		proxyEnvVarRef.SecretRef.Name = "guard-proxy"
		d.Spec.Template.Spec.Containers[0].EnvFrom = append(d.Spec.Template.Spec.Containers[0].EnvFrom, proxyEnvVarRef)
		if authopts.ProxyCert != "" {
			proxycertVolumeRef := core.Volume{
				Name: "proxy-certstore",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: "guard-proxy-cert",
					},
				},
			}

			sslCertsVolumeRef := core.Volume{
				Name: "ssl-certs",
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			}

			d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, proxycertVolumeRef)
			d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, sslCertsVolumeRef)

			sslCertsVolumeMount := core.VolumeMount{
				Name:      "ssl-certs",
				MountPath: "/etc/ssl/certs/",
			}

			d.Spec.Template.Spec.Containers[0].VolumeMounts = append(d.Spec.Template.Spec.Containers[0].VolumeMounts, sslCertsVolumeMount)

			optionalSecret := false
			initContainer := []core.Container{
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
							Name:      "proxy-certstore",
							MountPath: "/usr/local/share/ca-certificates/proxy-cert.crt",
							SubPath:   "proxy-cert.crt",
						},
						{
							Name:      "ssl-certs",
							MountPath: "/etc/ssl/certs/",
						},
					},
				},
			}

			d.Spec.Template.Spec.InitContainers = initContainer
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
		if extras, err := authopts.Azure.Apply(d); err != nil {
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

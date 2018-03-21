package installer

import (
	"fmt"

	stringz "github.com/appscode/go/strings"
	"github.com/appscode/go/types"
	v "github.com/appscode/go/version"
	"github.com/appscode/guard/server"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newDeployment(opts Options) (objects []runtime.Object, err error) {
	d := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: opts.namespace,
			Labels:    labels,
		},
		Spec: apps.DeploymentSpec{
			Replicas: types.Int32P(1),
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
				},
				Spec: core.PodSpec{
					ServiceAccountName: "guard",
					Containers: []core.Container{
						{
							Name:  "guard",
							Image: fmt.Sprintf("%s/guard:%v", opts.privateRegistry, stringz.Val(v.Version.Version, "canary")),
							Args: []string{
								"run",
								"--v=3",
							},
							Ports: []core.ContainerPort{
								{
									ContainerPort: server.ServingPort,
								},
							},
							ReadinessProbe: &core.Probe{
								Handler: core.Handler{
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
	if opts.imagePullSecret != "" {
		d.Spec.Template.Spec.ImagePullSecrets = []core.LocalObjectReference{
			{
				Name: opts.imagePullSecret,
			},
		}
	}
	if opts.runOnMaster {
		d.Spec.Template.Spec.NodeSelector = map[string]string{
			"node-role.kubernetes.io/master": "",
		}
		d.Spec.Template.Spec.Tolerations = append(d.Spec.Template.Spec.Tolerations, core.Toleration{
			Key:      "node-role.kubernetes.io/master",
			Operator: core.TolerationOpExists,
			Effect:   core.TaintEffectNoSchedule,
		})
	}
	objects = append(objects, d)

	servingOpts := server.NewSecureServingOptionsFromDir(opts.pkiDir)
	if extras, err := servingOpts.Apply(d); err != nil {
		return nil, err
	} else {
		objects = append(objects, extras...)
	}

	if extras, err := opts.Token.Apply(d); err != nil {
		return nil, err
	} else {
		objects = append(objects, extras...)
	}

	if extras, err := opts.Google.Apply(d); err != nil {
		return nil, err
	} else {
		objects = append(objects, extras...)
	}

	if extras, err := opts.Azure.Apply(d); err != nil {
		return nil, err
	} else {
		objects = append(objects, extras...)
	}

	if extras, err := opts.LDAP.Apply(d); err != nil {
		return nil, err
	} else {
		objects = append(objects, extras...)
	}

	return
}

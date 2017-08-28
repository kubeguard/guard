package cmds

import (
	"bytes"
	"fmt"

	stringz "github.com/appscode/go/strings"
	"github.com/appscode/go/types"
	v "github.com/appscode/go/version"
	"github.com/appscode/log"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	apps "k8s.io/client-go/pkg/apis/apps/v1beta1"
)

func NewCmdInstaller() *cobra.Command {
	var namespace string
	cmd := &cobra.Command{
		Use:               "installer",
		Short:             "Prints Kubernetes objects for deploying guard server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			store, err := NewCertStore()
			if err != nil {
				log.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if !store.PairExists("ca") {
				log.Fatalf("CA certificates not found in %s. Run `guard init ca`", store.Location())
			}
			if !store.PairExists("server") {
				log.Fatalf("Server certificate not found in %s. Run `guard init server`", store.Location())
			}

			caCert, _, err := store.ReadBytes("ca")
			if err != nil {
				log.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}
			serverCert, serverKey, err := store.ReadBytes("server")
			if err != nil {
				log.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}

			var buf bytes.Buffer
			secBytes, err := yaml.Marshal(createSecret(namespace, serverCert, serverKey, caCert))
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(secBytes)
			buf.WriteString("---\n")
			depBytes, err := yaml.Marshal(createDeployment(namespace))
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(depBytes)
			buf.WriteString("---\n")
			svcBytes, err := yaml.Marshal(createService(namespace))
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(svcBytes)

			fmt.Println(buf.String())
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "kube-system", "Name of Kubernetes namespace used to run guard server.")
	return cmd
}

var labels = map[string]string{
	"app": "guard",
}

func createSecret(namespace string, cert, key, caCert []byte) apiv1.Secret {
	return apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-pki",
			Namespace: namespace,
			Labels:    labels,
		},
		Data: map[string][]byte{
			"ca.crt":  caCert,
			"tls.crt": cert,
			"tls.key": key,
		},
	}
}

func createDeployment(namespace string) apps.Deployment {
	return apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: apps.DeploymentSpec{
			Replicas: types.Int32P(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "guard",
							Image: fmt.Sprintf("appscode/guard:%v", stringz.Val(v.Version.Version, "canary")),
							Args: []string{
								"run",
								"--v=3",
								"--ca-cert-file=/srv/guard/pki/ca.crt",
								"--cert-file=/srv/guard/pki/tls.crt",
								"--key-file=/srv/guard/pki/tls.key",
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "web",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: webPort,
								},
								{
									Name:          "ops",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: opsPort,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "guard-pki",
									MountPath: "/srv/guard/pki",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "guard-pki",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName:  "guard-pki",
									DefaultMode: types.Int32P(0555),
								},
							},
						},
					},
				},
			},
		},
	}
}

func createService(namespace string) apiv1.Service {
	return apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeClusterIP,
			Ports: []apiv1.ServicePort{
				{
					Name:       "web",
					Port:       webPort,
					Protocol:   apiv1.ProtocolTCP,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: labels,
		},
	}
}

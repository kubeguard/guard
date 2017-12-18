package cmds

import (
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/appscode/go/log"
	stringz "github.com/appscode/go/strings"
	"github.com/appscode/go/types"
	v "github.com/appscode/go/version"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewCmdInstaller() *cobra.Command {
	var (
		namespace  string
		addr       string
		enableRBAC bool
	)
	cmd := &cobra.Command{
		Use:               "installer",
		Short:             "Prints Kubernetes objects for deploying guard server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				log.Fatalf("Guard server address is invalid. Reason: %v.", err)
			}
			_, err = strconv.Atoi(port)
			if err != nil {
				log.Fatalf("Guard server port is invalid. Reason: %v.", err)
			}

			store, err := NewCertStore(rootDir)
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
			if enableRBAC {
				saBytes, err := yaml.Marshal(createServiceAccount(namespace))
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(saBytes)

				buf.WriteString("---\n")
				roleBytes, err := yaml.Marshal(createRole(namespace))
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(roleBytes)

				buf.WriteString("---\n")
				rbBytes, err := yaml.Marshal(createRoleBinding(namespace))
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(rbBytes)
			}

			buf.WriteString("---\n")
			secBytes, err := yaml.Marshal(createSecret(namespace, serverCert, serverKey, caCert))
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(secBytes)

			buf.WriteString("---\n")
			depBytes, err := yaml.Marshal(createDeployment(namespace, enableRBAC))
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(depBytes)

			buf.WriteString("---\n")
			svcBytes, err := yaml.Marshal(createService(namespace, addr))
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(svcBytes)

			fmt.Println(buf.String())
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "kube-system", "Name of Kubernetes namespace used to run guard server.")
	cmd.Flags().StringVar(&addr, "addr", "10.96.10.96:9844", "Address (host:port) of guard server.")
	cmd.Flags().BoolVar(&enableRBAC, "rbac", enableRBAC, "If true, uses RBAC with operator and database objects")
	return cmd
}

var labels = map[string]string{
	"app": "guard",
}

func createSecret(namespace string, cert, key, caCert []byte) core.Secret {
	return core.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: core.SchemeGroupVersion.String(),
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

func createDeployment(namespace string, enableRBAC bool) apps.Deployment {
	d := apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apps.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: apps.DeploymentSpec{
			Replicas: types.Int32P(1),
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
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
							Ports: []core.ContainerPort{
								{
									Name:          "web",
									Protocol:      core.ProtocolTCP,
									ContainerPort: webPort,
								},
								{
									Name:          "ops",
									Protocol:      core.ProtocolTCP,
									ContainerPort: opsPort,
								},
							},
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "guard-pki",
									MountPath: "/srv/guard/pki",
								},
							},
						},
					},
					Volumes: []core.Volume{
						{
							Name: "guard-pki",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
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
	if enableRBAC {
		d.Spec.Template.Spec.ServiceAccountName = "guard"
	}
	return d
}

func createService(namespace, addr string) core.Service {
	host, port, _ := net.SplitHostPort(addr)
	svcPort, _ := strconv.Atoi(port)
	return core.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: core.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: core.ServiceSpec{
			Type:      core.ServiceTypeClusterIP,
			ClusterIP: host,
			Ports: []core.ServicePort{
				{
					Name:       "web",
					Port:       int32(svcPort),
					Protocol:   core.ProtocolTCP,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: labels,
		},
	}
}

func createServiceAccount(namespace string) core.ServiceAccount {
	return core.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: core.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func createRole(namespace string) rbac.Role {
	return rbac.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbac.SchemeGroupVersion.String(),
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{core.GroupName},
				Resources: []string{"nodes"},
				Verbs:     []string{"list"},
			},
		},
	}
}

func createRoleBinding(namespace string) rbac.RoleBinding {
	return rbac.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbac.SchemeGroupVersion.String(),
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "Role",
			Name:     "guard",
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Name:      "guard",
				Namespace: namespace,
			},
		},
	}
}

package cmds

import (
	"bytes"
	"fmt"
	"net"
	"path/filepath"
	"strconv"

	"github.com/appscode/go/log"
	stringz "github.com/appscode/go/strings"
	"github.com/appscode/go/types"
	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

			store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(rootDir, "pki"))
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
			var data []byte

			if namespace != "kube-system" && namespace != core.NamespaceDefault {
				data, err = meta.MarshalToYAML(newNamespace(namespace), core.SchemeGroupVersion)
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(data)
				buf.WriteString("---\n")
			}

			if enableRBAC {
				data, err = meta.MarshalToYAML(newServiceAccount(namespace), core.SchemeGroupVersion)
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(data)
				buf.WriteString("---\n")

				data, err = meta.MarshalToYAML(newClusterRole(namespace), rbac.SchemeGroupVersion)
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(data)
				buf.WriteString("---\n")

				data, err = meta.MarshalToYAML(newClusterRoleBinding(namespace), rbac.SchemeGroupVersion)
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(data)
				buf.WriteString("---\n")
			}

			data, err = meta.MarshalToYAML(newSecret(namespace, serverCert, serverKey, caCert), core.SchemeGroupVersion)
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(data)
			buf.WriteString("---\n")

			data, err = meta.MarshalToYAML(newDeployment(namespace, enableRBAC), apps.SchemeGroupVersion)
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(data)
			buf.WriteString("---\n")

			data, err = meta.MarshalToYAML(newService(namespace, addr), core.SchemeGroupVersion)
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(data)

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

func newNamespace(namespace string) runtime.Object {
	return &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: labels,
		},
	}
}

func newSecret(namespace string, cert, key, caCert []byte) runtime.Object {
	return &core.Secret{
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

func newDeployment(namespace string, enableRBAC bool) runtime.Object {
	d := apps.Deployment{
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
								"--ca-cert-file=/etc/guard/pki/ca.crt",
								"--cert-file=/etc/guard/pki/tls.crt",
								"--key-file=/etc/guard/pki/tls.key",
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
									MountPath: "/etc/guard/pki",
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
	return &d
}

func newService(namespace, addr string) runtime.Object {
	host, port, _ := net.SplitHostPort(addr)
	svcPort, _ := strconv.Atoi(port)
	return &core.Service{
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

func newServiceAccount(namespace string) runtime.Object {
	return &core.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func newClusterRole(namespace string) runtime.Object {
	return &rbac.ClusterRole{
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

func newClusterRoleBinding(namespace string) runtime.Object {
	return &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
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

package cmds

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"

	"github.com/appscode/go/log"
	stringz "github.com/appscode/go/strings"
	"github.com/appscode/go/types"
	v "github.com/appscode/go/version"
	"github.com/appscode/guard/lib"
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

type options struct {
	namespace       string
	addr            string
	enableRBAC      bool
	runOnMaster     bool
	privateRegistry string
	imagePullSecret string
	tokenAuthFile   string
	Google          lib.GoogleOptions
	Azure           lib.AzureOptions
	LDAP            lib.LDAPOptions
}

func NewCmdInstaller() *cobra.Command {
	opts := options{
		namespace:       "kube-system",
		addr:            "10.96.10.96:443",
		privateRegistry: "appscode",
	}
	cmd := &cobra.Command{
		Use:               "installer",
		Short:             "Prints Kubernetes objects for deploying guard server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			_, port, err := net.SplitHostPort(opts.addr)
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

			if opts.namespace != "kube-system" && opts.namespace != core.NamespaceDefault {
				data, err = meta.MarshalToYAML(newNamespace(opts.namespace), core.SchemeGroupVersion)
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(data)
				buf.WriteString("---\n")
			}

			if opts.enableRBAC {
				data, err = meta.MarshalToYAML(newServiceAccount(opts.namespace), core.SchemeGroupVersion)
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(data)
				buf.WriteString("---\n")

				data, err = meta.MarshalToYAML(newClusterRole(opts.namespace), rbac.SchemeGroupVersion)
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(data)
				buf.WriteString("---\n")

				data, err = meta.MarshalToYAML(newClusterRoleBinding(opts.namespace), rbac.SchemeGroupVersion)
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(data)
				buf.WriteString("---\n")
			}

			data, err = meta.MarshalToYAML(newSecret(opts.namespace, serverCert, serverKey, caCert), core.SchemeGroupVersion)
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(data)
			buf.WriteString("---\n")

			secretData := map[string][]byte{}
			if opts.tokenAuthFile != "" {
				_, err := lib.LoadTokenFile(opts.tokenAuthFile)
				if err != nil {
					log.Fatalln(err)
				}
				tokens, err := ioutil.ReadFile(opts.tokenAuthFile)
				if err != nil {
					log.Fatalln(err)
				}
				secretData["token.csv"] = tokens
			}
			if opts.Google.ServiceAccountJsonFile != "" {
				sa, err := ioutil.ReadFile(opts.tokenAuthFile)
				if err != nil {
					log.Fatalln(err)
				}
				secretData["sa.json"] = sa
			}
			if len(secretData) > 0 {
				data, err = meta.MarshalToYAML(newSecretForTokenAuth(opts.namespace, secretData), core.SchemeGroupVersion)
				if err != nil {
					log.Fatalln(err)
				}
				buf.Write(data)
				buf.WriteString("---\n")
			}

			data, err = meta.MarshalToYAML(newDeployment(opts), apps.SchemeGroupVersion)
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(data)
			buf.WriteString("---\n")

			data, err = meta.MarshalToYAML(newService(opts.namespace, opts.addr), core.SchemeGroupVersion)
			if err != nil {
				log.Fatalln(err)
			}
			buf.Write(data)

			fmt.Println(buf.String())
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	cmd.Flags().StringVarP(&opts.namespace, "namespace", "n", opts.namespace, "Name of Kubernetes namespace used to run guard server.")
	cmd.Flags().StringVar(&opts.addr, "addr", opts.addr, "Address (host:port) of guard server.")
	cmd.Flags().BoolVar(&opts.enableRBAC, "rbac", opts.enableRBAC, "If true, uses RBAC with operator and database objects")
	cmd.Flags().BoolVar(&opts.runOnMaster, "run-on-master", opts.runOnMaster, "If true, runs Guard server on master instances")
	cmd.Flags().StringVar(&opts.privateRegistry, "private-registry", opts.privateRegistry, "Private Docker registry")
	cmd.Flags().StringVar(&opts.imagePullSecret, "image-pull-secret", opts.imagePullSecret, "Name of image pull secret")
	cmd.Flags().StringVar(&opts.tokenAuthFile, "token-auth-file", "", "Path to the token file")
	opts.Google.AddFlags(cmd.Flags())
	opts.Azure.AddFlags(cmd.Flags())
	opts.LDAP.AddFlags(cmd.Flags())
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

func newDeployment(opts options) runtime.Object {
	d := apps.Deployment{
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
					Containers: []core.Container{
						{
							Name:  "guard",
							Image: fmt.Sprintf("%s/guard:%v", opts.privateRegistry, stringz.Val(v.Version.Version, "canary")),
							Args: []string{
								"run",
								"--v=3",
								"--ca-cert-file=/etc/guard/pki/ca.crt",
								"--cert-file=/etc/guard/pki/tls.crt",
								"--key-file=/etc/guard/pki/tls.key",
							},
							Ports: []core.ContainerPort{
								{
									Name:          "api",
									Protocol:      core.ProtocolTCP,
									ContainerPort: port,
								},
							},
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "guard-pki",
									MountPath: "/etc/guard/pki",
								},
							},
							ReadinessProbe: &core.Probe{
								Handler: core.Handler{
									HTTPGet: &core.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(port),
										Scheme: core.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: int32(30),
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
	if opts.enableRBAC {
		d.Spec.Template.Spec.ServiceAccountName = "guard"
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

	if opts.tokenAuthFile != "" || opts.Google.ServiceAccountJsonFile != "" {
		volMount := core.VolumeMount{
			Name:      "guard-auth",
			MountPath: "/etc/guard/auth",
		}
		d.Spec.Template.Spec.Containers[0].VolumeMounts = append(d.Spec.Template.Spec.Containers[0].VolumeMounts, volMount)

		vol := core.Volume{
			Name: "guard-auth",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName:  "guard-auth",
					DefaultMode: types.Int32P(0555),
				},
			},
		}
		d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, vol)
	}

	if opts.tokenAuthFile != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, "--token-auth-file=/etc/guard/auth/token.csv")
	}
	d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, opts.Google.ToArgs()...)
	d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, opts.Azure.ToArgs()...)
	d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, opts.LDAP.ToArgs()...)

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
					Name:       "api",
					Port:       int32(svcPort),
					Protocol:   core.ProtocolTCP,
					TargetPort: intstr.FromString("api"),
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

func newSecretForTokenAuth(namespace string, data map[string][]byte) runtime.Object {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-auth",
			Namespace: namespace,
			Labels:    labels,
		},
		Data: data,
	}
}

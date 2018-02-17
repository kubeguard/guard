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
	namespace     string
	addr          string
	enableRBAC    bool
	tokenAuthFile string
	Azure         lib.AzureOptions
	Ldap          lib.LDAPOptions
}

func NewCmdInstaller() *cobra.Command {
	var opts options
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

			if opts.tokenAuthFile != "" {
				_, err := lib.LoadTokenFile(opts.tokenAuthFile)
				if err != nil {
					log.Fatalln(err)
				}
				tokenData, err := ioutil.ReadFile(opts.tokenAuthFile)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = meta.MarshalToYAML(newSecretForTokenAuth(opts.namespace, tokenData), core.SchemeGroupVersion)
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
	cmd.Flags().StringVarP(&opts.namespace, "namespace", "n", "kube-system", "Name of Kubernetes namespace used to run guard server.")
	cmd.Flags().StringVar(&opts.addr, "addr", "10.96.10.96:9844", "Address (host:port) of guard server.")
	cmd.Flags().BoolVar(&opts.enableRBAC, "rbac", opts.enableRBAC, "If true, uses RBAC with operator and database objects")
	cmd.Flags().StringVar(&opts.tokenAuthFile, "token-auth-file", "", "Path to the token file")
	opts.Azure.AddFlags(cmd.Flags())
	opts.Ldap.AddFlags(cmd.Flags())
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
	if opts.enableRBAC {
		d.Spec.Template.Spec.ServiceAccountName = "guard"
	}

	if opts.tokenAuthFile != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, "--token-auth-file=/etc/guard/auth/token.csv")
		volMount := core.VolumeMount{
			Name:      "guard-token-auth",
			MountPath: "/etc/guard/auth",
		}
		d.Spec.Template.Spec.Containers[0].VolumeMounts = append(d.Spec.Template.Spec.Containers[0].VolumeMounts, volMount)

		vol := core.Volume{
			Name: "guard-token-auth",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName:  "guard-token-auth",
					DefaultMode: types.Int32P(0555),
				},
			},
		}
		d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, vol)
	}

	//Add server flags
	if opts.Azure.ClientID != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--azure-client-id=%s", opts.Azure.ClientID))
	}
	if opts.Azure.ClientSecret != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--azure-client-secret=%s", opts.Azure.ClientSecret))
	}
	if opts.Azure.TenantID != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--azure-tenant-id=%s", opts.Azure.TenantID))
	}

	//Add server flags
	if opts.Ldap.ServerAddress != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-server-address=%s", opts.Ldap.ServerAddress))
	}
	if opts.Ldap.ServerPort != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-server-port=%s", opts.Ldap.ServerPort))
	}
	if opts.Ldap.BindDN != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-bind-dn=%s", opts.Ldap.BindDN))
	}
	if opts.Ldap.BindPassword != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-bind-password=%s", opts.Ldap.BindPassword))
	}
	if opts.Ldap.UserSearchDN != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-user-search-dn=%s", opts.Ldap.UserSearchDN))
	}
	if opts.Ldap.UserSearchFilter != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-user-search-filter=%s", opts.Ldap.UserSearchFilter))
	}
	if opts.Ldap.UserSearchFilter != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-user-attribute=%s", opts.Ldap.UserAttribute))
	}
	if opts.Ldap.GroupSearchDN != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-group-search-dn=%s", opts.Ldap.GroupSearchDN))
	}
	if opts.Ldap.GroupSearchFilter != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-group-search-filter=%s", opts.Ldap.GroupSearchFilter))
	}
	if opts.Ldap.GroupMemberAttribute != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-group-member-attribute=%s", opts.Ldap.GroupMemberAttribute))
	}
	if opts.Ldap.GroupNameAttribute != "" {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-group-name-attribute=%s", opts.Ldap.GroupNameAttribute))
	}
	if opts.Ldap.SkipTLSVerification {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-skip-tls-verification"))
	}
	if opts.Ldap.IsSecureLDAP {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-is-secure-ldap"))
	}
	if opts.Ldap.StartTLS {
		d.Spec.Template.Spec.Containers[0].Args = append(d.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("--ldap-start-tls"))
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

func newSecretForTokenAuth(namespace string, tokenFile []byte) runtime.Object {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-token-auth",
			Namespace: namespace,
			Labels:    labels,
		},
		Data: map[string][]byte{
			"token.csv": tokenFile,
		},
	}
}

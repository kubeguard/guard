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

package e2e_test

import (
	"context"
	"net"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/appscode/go/util/errors"
	"github.com/appscode/guard/auth/providers"
	"github.com/appscode/guard/auth/providers/azure"
	"github.com/appscode/guard/auth/providers/github"
	"github.com/appscode/guard/auth/providers/gitlab"
	"github.com/appscode/guard/auth/providers/google"
	"github.com/appscode/guard/auth/providers/ldap"
	"github.com/appscode/guard/auth/providers/token"
	"github.com/appscode/guard/installer"
	"github.com/appscode/guard/test/e2e/framework"

	"github.com/golang/glog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"gomodules.xyz/cert"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Installer test", func() {

	const (
		privateRegistryName    = "appscode"
		serverAddr             = "10.96.10.96"
		yamlDir                = "test-guard/yaml"
		certDir                = "test-guard/certs"
		tokenAuthDir           = "test-guard/auth"
		tokenFileName          = "token.csv"
		saDir                  = "test-guard/sa"
		saFileName             = "sa.json"
		installerfileName      = "installer.yaml"
		serviceName            = "guard"
		deploymentName         = "guard"
		clusterRoleName        = "guard"
		clusterRoleBindingName = "guard"
		pkiSecret              = "guard-pki"
		googleSecret           = "guard-google-auth"
		azureSecret            = "guard-azure-auth"
		ldapSecret             = "guard-ldap-auth"
		tokenSecret            = "guard-token-auth"
		timeOut                = 10 * time.Minute
		pollingInterval        = 10 * time.Second
		tokenData              = "token,username,uid,group"
		saData                 = `{
								   "type": "service_account",
								   "project_id": "",
								   "private_key_id": "",
								   "private_key": "",
								   "client_email": "c@g.com",
								   "client_id": "1",
								   "auth_uri": "https://accounts.google.com/o/oauth2/auth",
								   "token_uri": "https://accounts.google.com/o/oauth2/token",
								   "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
								   "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/who.iam.gserviceaccount.com"
								 }`
	)

	var (
		f     *framework.Invocation
		appFs afero.Fs
	)

	var (
		githubOpts = github.Options{
			BaseUrl: "url",
		}

		gitlabOpts = gitlab.Options{
			BaseUrl: "url",
		}
		azureOpts = azure.Options{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TenantID:     "tenant_id",
		}

		ldapOpts = ldap.Options{
			ServerAddress:        "host.com",
			ServerPort:           "389",
			BindDN:               "uid=admin,ou=system",
			BindPassword:         "secret",
			UserSearchDN:         "o=Company,ou=users",
			UserSearchFilter:     ldap.DefaultUserSearchFilter,
			UserAttribute:        ldap.DefaultUserAttribute,
			GroupSearchDN:        "o=Company,ou=groups",
			GroupSearchFilter:    ldap.DefaultGroupSearchFilter,
			GroupMemberAttribute: ldap.DefaultGroupMemberAttribute,
			GroupNameAttribute:   ldap.DefaultGroupNameAttribute,
			SkipTLSVerification:  true,
			StartTLS:             false,
			IsSecureLDAP:         false,
		}

		tokenOpts = token.Options{
			AuthFile: filepath.Join(tokenAuthDir, tokenFileName),
		}

		googleOpts = google.Options{
			ServiceAccountJsonFile: filepath.Join(saDir, saFileName),
			AdminEmail:             "admin@gmail.com",
		}
	)

	var (
		setupGuard = func(authopts installer.AuthOptions, authzopts installer.AuthzOptions) {
			By("Validate installer flag options")
			errs := authopts.Validate()
			Expect(errors.NewAggregate(errs)).NotTo(HaveOccurred())

			By("Generating installer yaml")
			data, err := installer.Generate(authopts, authzopts)
			Expect(err).NotTo(HaveOccurred())

			glog.Info(string(data))

			file := filepath.Join(yamlDir, installerfileName)
			By("Writing installer yaml in " + file)
			err = afero.WriteFile(appFs, file, data, 0777)
			Expect(err).NotTo(HaveOccurred())

			By("Executing command : kubectl apply -f " + file)
			cmd := "kubectl"
			args := []string{"apply", "-f", file}
			err = exec.Command(cmd, args...).Run()
			Expect(err).NotTo(HaveOccurred())
		}

		checkServiceCreated = func() {
			By("Checking service created. service name: " + serviceName)
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().Services(root.Namespace()).Get(context.TODO(), serviceName, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkDeploymentCreated = func() {
			By("Checking deployment created. deployment name: " + deploymentName)
			Eventually(func() bool {
				if obj, err := f.KubeClient.AppsV1().Deployments(root.Namespace()).Get(context.TODO(), deploymentName, metav1.GetOptions{}); err == nil {
					return *obj.Spec.Replicas == obj.Status.ReadyReplicas
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkPodCreated = func() {
			By("Checking pod created.")
			Eventually(func() bool {
				pods, err := f.KubeClient.CoreV1().Pods(root.Namespace()).List(context.TODO(), metav1.ListOptions{
					LabelSelector: "app=guard", // pods created for has has a label, app:guard
				})
				Expect(err).NotTo(HaveOccurred())
				return len(pods.Items) == 1
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkSecretCreated = func(name string) {
			By("Checking secret created. secret name: " + name)
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().Secrets(root.Namespace()).Get(context.TODO(), name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkClusterRoleCreated = func() {
			By("Checking cluster role created.")
			Eventually(func() bool {
				_, err := f.KubeClient.RbacV1().ClusterRoles().Get(context.TODO(), clusterRoleName, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkClusterRoleBindingCreated = func() {
			By("Checking cluster role binding created.")
			Eventually(func() bool {
				_, err := f.KubeClient.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkServiceDeleted = func() {
			By("Checking service Deleted. service name: " + serviceName)
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().Services(root.Namespace()).Get(context.TODO(), serviceName, metav1.GetOptions{})
				return kerr.IsNotFound(err) || kerr.IsGone(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkDeploymentDeleted = func() {
			By("Checking deployment Deleted. deployment name: " + deploymentName)
			Eventually(func() bool {
				_, err := f.KubeClient.AppsV1().Deployments(root.Namespace()).Get(context.TODO(), deploymentName, metav1.GetOptions{})
				return kerr.IsNotFound(err) || kerr.IsGone(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkPodDeleted = func() {
			By("Checking pod Deleted.")
			Eventually(func() bool {
				pods, err := f.KubeClient.CoreV1().Pods(root.Namespace()).List(context.TODO(), metav1.ListOptions{
					LabelSelector: "app=guard", // pods Deleted for has has a label, app:guard
				})
				Expect(err).NotTo(HaveOccurred())
				return len(pods.Items) == 0
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkSecretDeleted = func(name string) {
			By("Checking secret Deleted. secret name: " + name)
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().Secrets(root.Namespace()).Get(context.TODO(), name, metav1.GetOptions{})
				return kerr.IsNotFound(err) || kerr.IsGone(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkClusterRoleDeleted = func() {
			By("Checking cluster role Deleted.")
			Eventually(func() bool {
				_, err := f.KubeClient.RbacV1().ClusterRoles().Get(context.TODO(), clusterRoleName, metav1.GetOptions{})
				return kerr.IsNotFound(err) || kerr.IsGone(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkClusterRoleBindingDeleted = func() {
			By("Checking cluster role binding Deleted.")
			Eventually(func() bool {
				_, err := f.KubeClient.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
				return kerr.IsNotFound(err) || kerr.IsGone(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}
	)

	BeforeEach(func() {
		By("Setting up certificates")
		appFs = afero.NewOsFs()
		f = root.Invoke()
		err := appFs.MkdirAll(yamlDir, 0777)
		Expect(err).NotTo(HaveOccurred())

		err = appFs.MkdirAll(filepath.Join(certDir, "pki"), 0777)
		Expect(err).NotTo(HaveOccurred())

		// write ca
		err = afero.WriteFile(appFs, filepath.Join(certDir, "pki", "ca.crt"), f.CertStore.CACertBytes(), 0777)
		Expect(err).NotTo(HaveOccurred())

		err = afero.WriteFile(appFs, filepath.Join(certDir, "pki", "ca.key"), f.CertStore.CAKeyBytes(), 0777)
		Expect(err).NotTo(HaveOccurred())

		// write server cert, key
		srvCert, srvKey, err := f.CertStore.NewServerCertPairBytes(cert.AltNames{
			DNSNames: []string{"server"},
			IPs:      []net.IP{net.ParseIP(serverAddr)},
		})
		Expect(err).NotTo(HaveOccurred())
		err = afero.WriteFile(appFs, filepath.Join(certDir, "pki", "server.crt"), srvCert, 0777)
		Expect(err).NotTo(HaveOccurred())
		err = afero.WriteFile(appFs, filepath.Join(certDir, "pki", "server.key"), srvKey, 0777)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := appFs.RemoveAll("test-guard")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Set up guard for individual auth provider", func() {
		var (
			secretName string
			authopts   installer.AuthOptions
			authzopts  installer.AuthzOptions
		)

		BeforeEach(func() {
			authopts = installer.AuthOptions{
				PkiDir:          certDir,
				RunOnMaster:     false,
				Namespace:       root.Namespace(),
				Addr:            serverAddr + ":443",
				PrivateRegistry: privateRegistryName,
			}

			authzopts = installer.AuthzOptions{}

			secretName = pkiSecret

			checkServiceDeleted()
			checkDeploymentDeleted()
			checkClusterRoleBindingDeleted()
			checkClusterRoleDeleted()
			checkPodDeleted()
			checkSecretDeleted(secretName)

		})

		AfterEach(func() {
			Expect(f.DeleteService(serviceName, root.Namespace())).NotTo(HaveOccurred())
			Expect(f.DeleteDeployment(deploymentName, root.Namespace())).NotTo(HaveOccurred())
			Expect(f.DeleteClusterRole(clusterRoleName)).NotTo(HaveOccurred())
			Expect(f.DeleteClusterRoleBinding(clusterRoleBindingName)).NotTo(HaveOccurred())
			Expect(f.DeleteSecret(secretName, root.Namespace())).NotTo(HaveOccurred())
		})

		Context("Setting up guard for github", func() {
			BeforeEach(func() {
				authopts.AuthProvider = providers.AuthProviders{Providers: []string{github.OrgType}}
			})

			It("Set up guard for github should be successful", func() {
				setupGuard(authopts, authzopts)

				checkServiceCreated()
				checkClusterRoleCreated()
				checkClusterRoleBindingCreated()
				checkDeploymentCreated()
				checkPodCreated()
				checkSecretCreated(secretName)
			})

			It("Set up guard for github should be successful, provided base url", func() {
				authopts.Github = githubOpts

				setupGuard(authopts, authzopts)

				checkServiceCreated()
				checkClusterRoleCreated()
				checkClusterRoleBindingCreated()
				checkDeploymentCreated()
				checkPodCreated()
				checkSecretCreated(secretName)
			})
		})

		Context("Setting up guard for gitlab", func() {
			BeforeEach(func() {
				authopts.AuthProvider = providers.AuthProviders{Providers: []string{gitlab.OrgType}}
			})

			It("Set up guard for gitlab should be successful", func() {
				setupGuard(authopts, authzopts)

				checkServiceCreated()
				checkClusterRoleCreated()
				checkClusterRoleBindingCreated()
				checkDeploymentCreated()
				checkPodCreated()
				checkSecretCreated(secretName)
			})

			It("Set up guard for gitlab should be successful, provided base url", func() {
				authopts.Gitlab = gitlabOpts

				setupGuard(authopts, authzopts)

				checkServiceCreated()
				checkClusterRoleCreated()
				checkClusterRoleBindingCreated()
				checkDeploymentCreated()
				checkPodCreated()
				checkSecretCreated(secretName)
			})
		})

		Context("Setting up guard for azure", func() {
			BeforeEach(func() {
				authopts.AuthProvider = providers.AuthProviders{Providers: []string{azure.OrgType}}
				authopts.Azure = azureOpts

				checkSecretDeleted(azureSecret)
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(azureSecret, root.Namespace())).NotTo(HaveOccurred())
			})

			It("Set up guard for azure should be successful", func() {
				setupGuard(authopts, authzopts)

				checkServiceCreated()
				checkClusterRoleCreated()
				checkClusterRoleBindingCreated()
				checkDeploymentCreated()
				checkPodCreated()
				checkSecretCreated(secretName)
				checkSecretCreated(azureSecret)
				// time.Sleep(55 * time.Minute)
			})
		})

		Context("Setting up guard for LDAP", func() {
			BeforeEach(func() {
				authopts.AuthProvider = providers.AuthProviders{Providers: []string{ldap.OrgType}}
				authopts.LDAP = ldapOpts

				checkSecretDeleted(ldapSecret)
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(ldapSecret, root.Namespace())).NotTo(HaveOccurred())
			})

			It("Set up guard for LDAP should be successful", func() {
				setupGuard(authopts, authzopts)

				checkServiceCreated()
				checkClusterRoleCreated()
				checkClusterRoleBindingCreated()
				checkDeploymentCreated()
				checkPodCreated()
				checkSecretCreated(secretName)
				checkSecretCreated(ldapSecret)
				// time.Sleep(55 * time.Minute)
			})
		})

		Context("Setting up guard for token auth", func() {
			BeforeEach(func() {
				authopts.AuthProvider = providers.AuthProviders{Providers: []string{token.OrgType}}
				authopts.Token = tokenOpts

				err := appFs.Mkdir(tokenAuthDir, 0777)
				Expect(err).NotTo(HaveOccurred())
				err = afero.WriteFile(appFs, filepath.Join(tokenAuthDir, tokenFileName), []byte(tokenData), 0777)
				Expect(err).NotTo(HaveOccurred())

				checkSecretDeleted(tokenSecret)
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(tokenSecret, root.Namespace())).NotTo(HaveOccurred())
				err := appFs.RemoveAll(tokenAuthDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Set up guard for token auth should be successful", func() {
				setupGuard(authopts, authzopts)

				checkServiceCreated()
				checkClusterRoleCreated()
				checkClusterRoleBindingCreated()
				checkDeploymentCreated()
				checkPodCreated()
				checkSecretCreated(secretName)
				checkSecretCreated(tokenSecret)
				// time.Sleep(55 * time.Minute)
			})
		})

		Context("Setting up guard for google", func() {
			BeforeEach(func() {
				authopts.AuthProvider = providers.AuthProviders{Providers: []string{google.OrgType}}
				authopts.Google = googleOpts

				err := appFs.Mkdir(saDir, 0777)
				Expect(err).NotTo(HaveOccurred())

				err = afero.WriteFile(appFs, filepath.Join(saDir, saFileName), []byte(saData), 0777)
				Expect(err).NotTo(HaveOccurred())

				checkSecretDeleted(googleSecret)
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(googleSecret, root.Namespace())).NotTo(HaveOccurred())
				err := appFs.RemoveAll(saDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Set up guard for google should be successful", func() {
				setupGuard(authopts, authzopts)

				checkServiceCreated()
				checkClusterRoleCreated()
				checkClusterRoleBindingCreated()
				checkDeploymentCreated()
				checkPodCreated()
				checkSecretCreated(secretName)
				checkSecretCreated(googleSecret)
				// time.Sleep(55 * time.Minute)
			})
		})

	})

	Describe("Setting up guard for all providers", func() {
		var (
			secretNames []string
			authopts    installer.AuthOptions
			authzopts   installer.AuthzOptions
		)

		BeforeEach(func() {
			authopts = installer.AuthOptions{
				PkiDir:          certDir,
				RunOnMaster:     false,
				Namespace:       root.Namespace(),
				Addr:            serverAddr + ":443",
				PrivateRegistry: privateRegistryName,
				Azure:           azureOpts,
				LDAP:            ldapOpts,
				Token:           tokenOpts,
				Google:          googleOpts,
			}

			authopts.AuthProvider = providers.AuthProviders{Providers: []string{
				azure.OrgType,
				github.OrgType,
				gitlab.OrgType,
				google.OrgType,
				ldap.OrgType,
				token.OrgType,
			},
			}

			secretNames = []string{
				pkiSecret,
				azureSecret,
				ldapSecret,
				tokenSecret,
				googleSecret,
			}

			err := appFs.Mkdir(tokenAuthDir, 0777)
			Expect(err).NotTo(HaveOccurred())
			err = afero.WriteFile(appFs, filepath.Join(tokenAuthDir, tokenFileName), []byte(tokenData), 0777)
			Expect(err).NotTo(HaveOccurred())

			err = appFs.Mkdir(saDir, 0777)
			Expect(err).NotTo(HaveOccurred())
			err = afero.WriteFile(appFs, filepath.Join(saDir, saFileName), []byte(saData), 0777)
			Expect(err).NotTo(HaveOccurred())

			checkServiceDeleted()
			checkDeploymentDeleted()
			checkClusterRoleBindingDeleted()
			checkClusterRoleDeleted()
			checkPodDeleted()

			for _, name := range secretNames {
				checkSecretDeleted(name)
			}
		})

		AfterEach(func() {
			Expect(f.DeleteService(serviceName, root.Namespace())).NotTo(HaveOccurred())
			Expect(f.DeleteDeployment(deploymentName, root.Namespace())).NotTo(HaveOccurred())
			Expect(f.DeleteClusterRole(clusterRoleName)).NotTo(HaveOccurred())
			Expect(f.DeleteClusterRoleBinding(clusterRoleBindingName)).NotTo(HaveOccurred())

			for _, name := range secretNames {
				Expect(f.DeleteSecret(name, root.Namespace())).NotTo(HaveOccurred())
			}
		})

		It("Set up guard for all providers should be successful", func() {
			setupGuard(authopts, authzopts)

			checkServiceCreated()
			checkClusterRoleCreated()
			checkClusterRoleBindingCreated()
			checkDeploymentCreated()
			checkPodCreated()

			for _, name := range secretNames {
				checkSecretCreated(name)
			}

		})
	})

})

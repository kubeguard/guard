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
	"os"
	"path/filepath"
	"testing"

	"go.kubeguard.dev/guard/auth/providers"
	"go.kubeguard.dev/guard/auth/providers/azure"

	"github.com/stretchr/testify/assert"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestAuthOptionsValidateWithAzureEntraSDK(t *testing.T) {
	t.Run("requires azure auth provider", func(t *testing.T) {
		authopts := NewAuthOptions()
		authopts.AuthProvider = providers.AuthProviders{Providers: []string{"github"}}
		authopts.UseAzureEntraSDK = true

		err := aggregateErrors(authopts.Validate())
		assert.ErrorContains(t, err, "azure.use-entra-sdk requires azure auth provider")
	})

	t.Run("requires a sidecar image when enabled", func(t *testing.T) {
		authopts := NewAuthOptions()
		authopts.AuthProvider = providers.AuthProviders{Providers: []string{azure.OrgType}}
		authopts.Azure = azure.Options{ClientID: "client-id", ClientSecret: "client-secret", TenantID: "tenant-id", AuthMode: azure.ClientCredentialAuthMode}
		authopts.UseAzureEntraSDK = true
		authopts.AzureEntraSDKImage = ""

		err := aggregateErrors(authopts.Validate())
		assert.EqualError(t, err, "azure.entra-sdk-image must be non-empty when azure.use-entra-sdk is enabled")
	})

	t.Run("installer Entra SDK flag overrides azure Entra SDK URL validation", func(t *testing.T) {
		authopts := NewAuthOptions()
		authopts.AuthProvider = providers.AuthProviders{Providers: []string{azure.OrgType}}
		authopts.Azure = azure.Options{ClientID: "client-id", ClientSecret: "client-secret", TenantID: "tenant-id", AuthMode: azure.ClientCredentialAuthMode, EntraSDKURL: "://bad-url"}
		authopts.UseAzureEntraSDK = true

		assert.NoError(t, aggregateErrors(authopts.Validate()))
	})
}

func TestNewDeploymentWithAzureEntraSDK(t *testing.T) {
	t.Run("adds Entra SDK sidecar and derived Guard URL", func(t *testing.T) {
		authopts := newAzureAuthOptions(t)
		authopts.UseAzureEntraSDK = true

		objects, err := newDeployment(authopts, AuthzOptions{})
		if !assert.NoError(t, err) {
			return
		}

		deployment := findDeployment(t, objects)
		if !assert.NotNil(t, deployment) {
			return
		}

		if assert.Len(t, deployment.Spec.Template.Spec.Containers, 2) {
			guardContainer := deployment.Spec.Template.Spec.Containers[0]
			entraSDKContainer := deployment.Spec.Template.Spec.Containers[1]

			assert.Equal(t, "guard", guardContainer.Name)
			assert.Contains(t, guardContainer.Args, "--azure.entra-sdk-url=http://127.0.0.1:8080")

			assert.Equal(t, azureEntraSDKContainerName, entraSDKContainer.Name)
			assert.Equal(t, DefaultAzureEntraSDKImage, entraSDKContainer.Image)
			assertEntraSDKEnvVars(t, entraSDKContainer.Env)
			assert.Empty(t, entraSDKContainer.Ports)
			assertEntraSDKProbe(t, entraSDKContainer.ReadinessProbe)
			assertEntraSDKProbe(t, entraSDKContainer.LivenessProbe)
		}
	})

	t.Run("installer flag overrides explicit azure Entra SDK URL", func(t *testing.T) {
		authopts := newAzureAuthOptions(t)
		authopts.UseAzureEntraSDK = true
		authopts.Azure.EntraSDKURL = "http://external-sdk.example.com"

		objects, err := newDeployment(authopts, AuthzOptions{})
		if !assert.NoError(t, err) {
			return
		}

		deployment := findDeployment(t, objects)
		if !assert.NotNil(t, deployment) {
			return
		}

		guardContainer := deployment.Spec.Template.Spec.Containers[0]
		assert.Contains(t, guardContainer.Args, "--azure.entra-sdk-url=http://127.0.0.1:8080")
		assert.NotContains(t, guardContainer.Args, "--azure.entra-sdk-url=http://external-sdk.example.com")
	})

	t.Run("uses custom Entra SDK image", func(t *testing.T) {
		authopts := newAzureAuthOptions(t)
		authopts.UseAzureEntraSDK = true
		authopts.AzureEntraSDKImage = "example.com/custom/entra-sdk:test"

		objects, err := newDeployment(authopts, AuthzOptions{})
		if !assert.NoError(t, err) {
			return
		}

		deployment := findDeployment(t, objects)
		if !assert.NotNil(t, deployment) {
			return
		}

		if assert.Len(t, deployment.Spec.Template.Spec.Containers, 2) {
			assert.Equal(t, "example.com/custom/entra-sdk:test", deployment.Spec.Template.Spec.Containers[1].Image)
		}
	})

	t.Run("uses local custom guard image", func(t *testing.T) {
		authopts := newAzureAuthOptions(t)
		authopts.PrivateRegistry = "appscode"
		authopts.GuardImage = "localhost/guard-e2e/guard:local"

		objects, err := newDeployment(authopts, AuthzOptions{})
		if !assert.NoError(t, err) {
			return
		}

		deployment := findDeployment(t, objects)
		if !assert.NotNil(t, deployment) {
			return
		}

		guardContainer := deployment.Spec.Template.Spec.Containers[0]
		assert.Equal(t, "localhost/guard-e2e/guard:local", guardContainer.Image)
		assert.Equal(t, core.PullNever, guardContainer.ImagePullPolicy)
	})

	t.Run("uses external custom guard image without forcing PullNever", func(t *testing.T) {
		authopts := newAzureAuthOptions(t)
		authopts.GuardImage = "ghcr.io/example/guard:test"

		objects, err := newDeployment(authopts, AuthzOptions{})
		if !assert.NoError(t, err) {
			return
		}

		deployment := findDeployment(t, objects)
		if !assert.NotNil(t, deployment) {
			return
		}

		guardContainer := deployment.Spec.Template.Spec.Containers[0]
		assert.Equal(t, "ghcr.io/example/guard:test", guardContainer.Image)
		assert.Empty(t, guardContainer.ImagePullPolicy)
	})

	t.Run("defaults verbosity when unset", func(t *testing.T) {
		authopts := newAzureAuthOptions(t)
		authopts.VerbosityLevel = ""

		objects, err := newDeployment(authopts, AuthzOptions{})
		if !assert.NoError(t, err) {
			return
		}

		deployment := findDeployment(t, objects)
		if !assert.NotNil(t, deployment) {
			return
		}

		guardContainer := deployment.Spec.Template.Spec.Containers[0]
		assert.Contains(t, guardContainer.Args, "--v=3")
		assert.NotContains(t, guardContainer.Args, "--v=")
	})
}

func newAzureAuthOptions(t *testing.T) AuthOptions {
	t.Helper()
	authopts := NewAuthOptions()
	authopts.RunOnMaster = false
	authopts.PkiDir = newTestPKIDir(t)
	authopts.AuthProvider = providers.AuthProviders{Providers: []string{azure.OrgType}}
	authopts.Azure = azure.Options{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TenantID:     "tenant-id",
		AuthMode:     azure.ClientCredentialAuthMode,
	}
	return authopts
}

func newTestPKIDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	pkiDir := filepath.Join(dir, "pki")
	if err := os.MkdirAll(pkiDir, 0o755); err != nil {
		t.Fatalf("failed to create test pki dir: %v", err)
	}
	for name, data := range map[string]string{
		"ca.crt":     "test-ca-crt",
		"ca.key":     "test-ca-key",
		"server.crt": "test-server-crt",
		"server.key": "test-server-key",
	} {
		if err := os.WriteFile(filepath.Join(pkiDir, name), []byte(data), 0o600); err != nil {
			t.Fatalf("failed to write test PKI file %s: %v", name, err)
		}
	}
	return dir
}

func findDeployment(t *testing.T, objects []runtime.Object) *apps.Deployment {
	t.Helper()
	for _, obj := range objects {
		if deployment, ok := obj.(*apps.Deployment); ok {
			return deployment
		}
	}
	t.Fatal("deployment not found")
	return nil
}

func aggregateErrors(errs []error) error {
	return utilerrors.NewAggregate(errs)
}

func assertEntraSDKProbe(t *testing.T, probe *core.Probe) {
	t.Helper()
	if assert.NotNil(t, probe) && assert.NotNil(t, probe.HTTPGet) {
		assert.Equal(t, "/healthz", probe.HTTPGet.Path)
		assert.Equal(t, intstr.FromInt(azureEntraSDKPort), probe.HTTPGet.Port)
		assert.Equal(t, core.URISchemeHTTP, probe.HTTPGet.Scheme)
		assert.Equal(t, []core.HTTPHeader{{
			Name:  "Host",
			Value: "localhost",
		}}, probe.HTTPGet.HTTPHeaders)
	}
}

func assertEntraSDKEnvVars(t *testing.T, envVars []core.EnvVar) {
	t.Helper()
	if assert.Len(t, envVars, 4) {
		assert.Equal(t, core.EnvVar{Name: "AzureAd__Instance", Value: "https://login.microsoftonline.com/"}, envVars[0])
		assert.Equal(t, core.EnvVar{Name: "AzureAd__TenantId", Value: "tenant-id"}, envVars[1])
		assert.Equal(t, core.EnvVar{Name: "AzureAd__ClientId", Value: "client-id"}, envVars[2])
		assert.Equal(t, core.EnvVar{Name: "AzureAd__Audience", Value: "client-id"}, envVars[3])
	}
}

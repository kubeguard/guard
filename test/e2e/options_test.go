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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.kubeguard.dev/guard/auth/providers/azure"
	"gomodules.xyz/logs"
	"k8s.io/client-go/util/homedir"
)

type E2EOptions struct {
	KubeContext       string
	KubeConfig        string
	GuardImage        string
	AzureEntraSDKAuth AzureEntraSDKE2EOptions
}

type AzureEntraSDKE2EOptions struct {
	Environment string
	ClientID    string
	TenantID    string
	AccessToken string
}

var options = &E2EOptions{
	KubeConfig:        filepath.Join(homedir.HomeDir(), ".kube", "config"),
	KubeContext:       "minikube",
	AzureEntraSDKAuth: loadAzureEntraSDKE2EOptions(),
}

func init() {
	flag.StringVar(&options.KubeConfig, "kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&options.KubeContext, "kube-context", "", "Name of kube context")
	flag.StringVar(&options.GuardImage, "guard-image", "", "Full Guard image reference to use in E2E deployment generation")
	enableLogging()
}

func enableLogging() {
	defer func() {
		logs.InitLogs()
		defer logs.FlushLogs()
	}()
	// err := flag.Set("logtostderr", "true")
	// if err != nil {
	//	klog.Fatal(err)
	// }
	logLevelFlag := flag.Lookup("v")
	if logLevelFlag != nil {
		if len(logLevelFlag.Value.String()) > 0 && logLevelFlag.Value.String() != "0" {
			return
		}
	}
	_ = flag.Set("v", strconv.Itoa(2))
}

func loadAzureEntraSDKE2EOptions() AzureEntraSDKE2EOptions {
	return AzureEntraSDKE2EOptions{
		Environment: strings.TrimSpace(os.Getenv("AZURE_E2E_ENVIRONMENT")),
		ClientID:    strings.TrimSpace(os.Getenv("AZURE_E2E_CLIENT_ID")),
		TenantID:    strings.TrimSpace(os.Getenv("AZURE_E2E_TENANT_ID")),
		AccessToken: strings.TrimSpace(os.Getenv("AZURE_E2E_ACCESS_TOKEN")),
	}
}

func (o AzureEntraSDKE2EOptions) Configured() bool {
	return len(o.Missing()) == 0
}

func (o AzureEntraSDKE2EOptions) Missing() []string {
	var missing []string
	if o.ClientID == "" {
		missing = append(missing, "AZURE_E2E_CLIENT_ID")
	}
	if o.TenantID == "" {
		missing = append(missing, "AZURE_E2E_TENANT_ID")
	}
	if o.AccessToken == "" {
		missing = append(missing, "AZURE_E2E_ACCESS_TOKEN")
	}
	return missing
}

func (o AzureEntraSDKE2EOptions) SkipMessage() string {
	return fmt.Sprintf("Azure Entra SDK E2E is not configured; missing %s", strings.Join(o.Missing(), ", "))
}

func (o AzureEntraSDKE2EOptions) AzureOptions() azure.Options {
	return azure.Options{
		Environment: o.Environment,
		ClientID:    o.ClientID,
		TenantID:    o.TenantID,
		// Use passthrough so Guard does not require Graph client credentials for
		// this verifier-focused E2E. The test only exercises access token
		// validation plus Guard's post-validation claim checks, and explicitly
		// skips group membership resolution.
		AuthMode:                                 azure.PassthroughAuthMode,
		ResolveGroupMembershipOnlyOnOverageClaim: true,
		SkipGroupMembershipResolution:            true,
		VerifyClientID:                           true,
	}
}

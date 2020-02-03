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
package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/appscode/guard/auth"

	"github.com/golang/glog"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func NewCmdGetWebhookConfig() *cobra.Command {
	var (
		rootDir = auth.DefaultDataDir
		org     string
		addr    string
	)
	cmd := &cobra.Command{
		Use:               "webhook-config",
		Short:             "Prints authentication token webhook config file",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			org = strings.ToLower(org)
			if len(args) == 0 {
				switch org {
				// for gitlab/azure/ldap/firebase client name not required
				case "gitlab", "azure", "ldap", "firebase":
					args = []string{org}
				}
			}

			if len(args) == 0 {
				glog.Fatalln("Missing client name.")
			}
			if len(args) > 1 {
				glog.Fatalln("Multiple client name found.")
			}

			cfg := cert.Config{
				AltNames: cert.AltNames{
					DNSNames: []string{args[0]},
				},
			}
			if org == "" {
				glog.Fatalf("Missing organization name. Set flag -o %s", auth.SupportedOrgs)
			} else if !auth.SupportedOrgs.Has(org) {
				glog.Fatalf("Unknown organization %s.", org)
			} else {
				cfg.Organization = []string{org}
			}

			store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(rootDir, "pki"))
			if err != nil {
				glog.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if !store.PairExists("ca") {
				glog.Fatalf("CA certificates not found in %s. Run `guard init ca`", store.Location())
			}
			if !store.PairExists(filename(cfg)) {
				glog.Fatalf("Client certificate not found in %s. Run `guard init client %s -o %s`", store.Location(), cfg.AltNames.DNSNames[0], cfg.Organization[0])
			}

			caCert, _, err := store.ReadBytes("ca")
			if err != nil {
				glog.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}
			clientCert, clientKey, err := store.ReadBytes(filename(cfg))
			if err != nil {
				glog.Fatalf("Failed to load client certificate. Reason: %v.", err)
			}

			config := clientcmdapi.Config{
				Kind:       "Config",
				APIVersion: "v1",
				Clusters: map[string]*clientcmdapi.Cluster{
					"guard-server": {
						Server:                   fmt.Sprintf("https://%s/tokenreviews", addr),
						CertificateAuthorityData: caCert,
					},
				},
				AuthInfos: map[string]*clientcmdapi.AuthInfo{
					filename(cfg): {
						ClientCertificateData: clientCert,
						ClientKeyData:         clientKey,
					},
				},
				Contexts: map[string]*clientcmdapi.Context{
					"webhook": {
						Cluster:  "guard-server",
						AuthInfo: filename(cfg),
					},
				},
				CurrentContext: "webhook",
			}
			data, err := clientcmd.Write(config)
			if err != nil {
				glog.Fatalln(err)
			}
			fmt.Println(string(data))
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	cmd.Flags().StringVarP(&org, "organization", "o", org, fmt.Sprintf("Name of Organization (%v).", auth.SupportedOrgs))
	cmd.Flags().StringVar(&addr, "addr", "10.96.10.96:443", "Address (host:port) of guard server.")
	return cmd
}

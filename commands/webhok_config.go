package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/appscode/guard/auth"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/golang/glog"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/cert"
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
				// for gitlab/azure/ldap client name not required
				case "gitlab", "azure", "ldap":
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
				CommonName: args[0],
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
				glog.Fatalf("Client certificate not found in %s. Run `guard init client %s -p %s`", store.Location(), cfg.CommonName, cfg.Organization[0])
			}

			caCert, _, err := store.ReadBytes("ca")
			if err != nil {
				glog.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}
			clientCert, clientKey, err := store.ReadBytes(filename(cfg))
			if err != nil {
				glog.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}

			config := clientcmdapi.Config{
				Kind:       "Config",
				APIVersion: "v1",
				Clusters: map[string]*clientcmdapi.Cluster{
					"guard-server": {
						Server: fmt.Sprintf("https://%s/tokenreviews", addr),
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

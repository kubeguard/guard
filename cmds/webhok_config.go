package cmds

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/ghodss/yaml"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	cli "k8s.io/client-go/tools/clientcmd/api/v1"
	"k8s.io/client-go/util/cert"
)

func NewCmdGetWebhookConfig() *cobra.Command {
	var org, addr string
	cmd := &cobra.Command{
		Use:               "webhook-config",
		Short:             "Prints authentication token webhook config file",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatalln("Missing client name.")
			}
			if len(args) > 1 {
				log.Fatalln("Multiple client name found.")
			}

			cfg := cert.Config{
				CommonName: args[0],
			}
			org = strings.ToLower(org)
			switch org {
			case "github":
				cfg.Organization = []string{"Github"}
			case "google":
				cfg.Organization = []string{"Google"}
			case "appscode":
				cfg.Organization = []string{"Appscode"}
			case "":
				log.Fatalln("Missing organization name. Set flag -o Google|Github.")
			default:
				log.Fatalf("Unknown organization %s.", org)
			}

			store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(rootDir, "pki"))
			if err != nil {
				log.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if !store.PairExists("ca") {
				log.Fatalf("CA certificates not found in %s. Run `guard init ca`", store.Location())
			}
			if !store.PairExists(filename(cfg)) {
				log.Fatalf("Client certificate not found in %s. Run `guard init client %s -p %s`", store.Location(), cfg.CommonName, cfg.Organization[0])
			}

			caCert, _, err := store.ReadBytes("ca")
			if err != nil {
				log.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}
			clientCert, clientKey, err := store.ReadBytes(filename(cfg))
			if err != nil {
				log.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}

			kubeconfig := cli.Config{
				Kind:       "Config",
				APIVersion: "v1",
				Clusters: []cli.NamedCluster{
					{
						Name: "guard-server",
						Cluster: cli.Cluster{
							Server: fmt.Sprintf("https://%s/apis/authentication.k8s.io/v1beta1/tokenreviews", addr),
							CertificateAuthorityData: caCert,
						},
					},
				},
				AuthInfos: []cli.NamedAuthInfo{
					{
						Name: filename(cfg),
						AuthInfo: cli.AuthInfo{
							ClientCertificateData: clientCert,
							ClientKeyData:         clientKey,
						},
					},
				},
				Contexts: []cli.NamedContext{
					{
						Name: "webhook",
						Context: cli.Context{
							Cluster:  "guard-server",
							AuthInfo: filename(cfg),
						},
					},
				},
				CurrentContext: "webhook",
			}
			bytes, err := yaml.Marshal(kubeconfig)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(string(bytes))
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	cmd.Flags().StringVarP(&org, "organization", "o", org, "Name of Organization (Github or Google).")
	cmd.Flags().StringVar(&addr, "addr", "10.96.10.96:9844", "Address (host:port) of guard server.")
	return cmd
}

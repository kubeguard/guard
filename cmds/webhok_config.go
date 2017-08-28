package cmds

import (
	"fmt"
	"strings"

	"github.com/appscode/log"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	cli "k8s.io/client-go/tools/clientcmd/api/v1"
	"k8s.io/client-go/util/cert"
)

func NewCmdGetWebhookConfig() *cobra.Command {
	var org string
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
			case "":
				log.Fatalln("Missing organization name. Set flag -o Google|Github.")
			default:
				log.Fatalf("Unknown organization %s.", org)
			}

			store, err := NewCertStore()
			if err != nil {
				log.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if !store.PairExists("ca") {
				log.Fatalf("CA certificates not found in %s. Run `guard init ca`", store.Location())
			}
			if !store.PairExists(store.Filename(cfg)) {
				log.Fatalf("Client certificate not found in %s. Run `guard init client %s -p %s`", store.Location(), cfg.CommonName, cfg.Organization[0])
			}

			caCert, _, err := store.ReadBytes("ca")
			if err != nil {
				log.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}
			clientCert, clientKey, err := store.ReadBytes(store.Filename(cfg))
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
							Server: "http(s)://>guard-server-host:port>/apis/authentication.k8s.io/v1beta1/tokenreviews",
							CertificateAuthorityData: caCert,
						},
					},
				},
				AuthInfos: []cli.NamedAuthInfo{
					{
						Name: store.Filename(cfg),
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
							AuthInfo: store.Filename(cfg),
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

	cmd.Flags().StringVarP(&org, "organization", "o", org, "Name of Organization (Github or Google).")
	return cmd
}

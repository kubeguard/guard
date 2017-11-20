package cmds

import (
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/appscode/go/log"
	"github.com/appscode/go/term"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/cert"
)

func NewCmdInitClient() *cobra.Command {
	var org string
	cmd := &cobra.Command{
		Use:               "client",
		Short:             "Generate client certificate pair",
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
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
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

			store, err := NewCertStore(rootDir)
			if err != nil {
				log.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if store.IsExists(store.Filename(cfg)) {
				if !term.Ask(fmt.Sprintf("Client certificate found at %s. Do you want to overwrite?", store.Location()), false) {
					os.Exit(1)
				}
			}

			if !store.PairExists("ca") {
				log.Fatalf("CA certificates not found in %s. Run `guard init ca`", store.Location())
			}
			caCert, caKey, err := store.Read("ca")
			if err != nil {
				log.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}

			key, err := cert.NewPrivateKey()
			if err != nil {
				log.Fatalf("Failed to generate private key. Reason: %v.", err)
			}
			cert, err := cert.NewSignedCert(cfg, key, caCert, caKey)
			if err != nil {
				log.Fatalf("Failed to generate server certificate. Reason: %v.", err)
			}
			err = store.Write(store.Filename(cfg), cert, key)
			if err != nil {
				log.Fatalf("Failed to init client certificate pair. Reason: %v.", err)
			}
			term.Successln("Wrote client certificates in ", store.Location())
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	cmd.Flags().StringVarP(&org, "organization", "o", org, "Name of Organization (Github or Google).")
	return cmd
}

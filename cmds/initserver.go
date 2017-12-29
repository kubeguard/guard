package cmds

import (
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/appscode/go/log"
	"github.com/appscode/go/term"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/cert"
)

func NewCmdInitServer() *cobra.Command {
	sans := cert.AltNames{
		IPs: []net.IP{net.ParseIP("127.0.0.1")},
	}
	cmd := &cobra.Command{
		Use:               "server",
		Short:             "Generate server certificate pair",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cert.Config{
				CommonName: "server",
				AltNames:   sans,
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			}

			store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(rootDir, "pki"), cfg.Organization...)
			if err != nil {
				log.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if store.IsExists(filename(cfg)) {
				if !term.Ask(fmt.Sprintf("Server certificate found at %s. Do you want to overwrite?", store.Location()), false) {
					os.Exit(1)
				}
			}

			if err = store.LoadCA(); err != nil {
				log.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}

			crt, key, err := store.NewServerCertPair(cfg.CommonName, cfg.AltNames)
			if err != nil {
				log.Fatalf("Failed to generate certificate pair. Reason: %v.", err)
			}
			err = store.WriteBytes(filename(cfg), crt, key)
			if err != nil {
				log.Fatalf("Failed to init server certificate pair. Reason: %v.", err)
			}
			term.Successln("Wrote server certificates in ", store.Location())
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	cmd.Flags().IPSliceVar(&sans.IPs, "ips", sans.IPs, "Alternative IP addresses")
	cmd.Flags().StringSliceVar(&sans.DNSNames, "domains", sans.DNSNames, "Alternative Domain names")
	return cmd
}

package cmds

import (
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/appscode/go-term"
	"github.com/appscode/log"
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

			store, err := NewCertStore()
			if err != nil {
				log.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if store.IsExists(store.Filename(cfg)) {
				if !term.Ask(fmt.Sprintf("Server certificate found at %s. Do you want to overwrite?", store.Location()), false) {
					os.Exit(1)
				}
			}

			if !store.PairExists("ca") {
				log.Fatalf("CA certificates not found in %s. Run `kad init ca`", store.Location())
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
				log.Fatalf("Failed to init server certificate pair. Reason: %v.", err)
			}
		},
	}

	cmd.Flags().IPSliceVar(&sans.IPs, "ips", sans.IPs, "Alternative IP addresses")
	cmd.Flags().StringSliceVar(&sans.DNSNames, "domains", sans.DNSNames, "Alternative Domain names")
	return cmd
}

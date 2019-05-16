package commands

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/appscode/go/term"
	"github.com/appscode/guard/auth"
	"github.com/golang/glog"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
)

func NewCmdInitClient() *cobra.Command {
	var (
		rootDir = auth.DefaultDataDir
		org     string
	)
	cmd := &cobra.Command{
		Use:               "client",
		Short:             "Generate client certificate pair",
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
				AltNames: cert.AltNames{
					DNSNames: []string{args[0]},
				},
				Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			if org == "" {
				glog.Fatalf("Missing organization name. Set flag -o %s", auth.SupportedOrgs)
			} else if !auth.SupportedOrgs.Has(org) {
				glog.Fatalf("Unknown organization %s.", org)
			} else {
				cfg.Organization = []string{org}
			}

			store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(rootDir, "pki"), cfg.Organization...)
			if err != nil {
				glog.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if store.IsExists(filename(cfg)) {
				if !term.Ask(fmt.Sprintf("Client certificate found at %s. Do you want to overwrite?", store.Location()), false) {
					os.Exit(1)
				}
			}

			if err = store.LoadCA(); err != nil {
				glog.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}

			crt, key, err := store.NewClientCertPairBytes(cfg.AltNames, cfg.Organization...)
			if err != nil {
				glog.Fatalf("Failed to generate certificate pair. Reason: %v.", err)
			}
			err = store.WriteBytes(filename(cfg), crt, key)
			if err != nil {
				glog.Fatalf("Failed to init client certificate pair. Reason: %v.", err)
			}
			term.Successln("Wrote client certificates in ", store.Location())
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	cmd.Flags().StringVarP(&org, "organization", "o", org, fmt.Sprintf("Name of Organization (%v).", auth.SupportedOrgs))
	return cmd
}

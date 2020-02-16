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
	"crypto/x509"
	"fmt"
	"net"
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

func NewCmdInitServer() *cobra.Command {
	var (
		rootDir = auth.DefaultDataDir
		sans    = cert.AltNames{
			IPs: []net.IP{net.ParseIP("127.0.0.1")},
		}
	)
	cmd := &cobra.Command{
		Use:               "server",
		Short:             "Generate server certificate pair",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cert.Config{
				AltNames: cert.AltNames{
					DNSNames: merge("server", sans.DNSNames),
					IPs:      sans.IPs,
				},
				Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			}

			store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(rootDir, "pki"), cfg.Organization...)
			if err != nil {
				glog.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if store.IsExists(filename(cfg)) {
				if !term.Ask(fmt.Sprintf("Server certificate found at %s. Do you want to overwrite?", store.Location()), false) {
					os.Exit(1)
				}
			}

			if err = store.LoadCA(); err != nil {
				glog.Fatalf("Failed to load ca certificate. Reason: %v.", err)
			}

			crt, key, err := store.NewServerCertPairBytes(cfg.AltNames)
			if err != nil {
				glog.Fatalf("Failed to generate certificate pair. Reason: %v.", err)
			}
			err = store.WriteBytes(filename(cfg), crt, key)
			if err != nil {
				glog.Fatalf("Failed to init server certificate pair. Reason: %v.", err)
			}
			term.Successln("Wrote server certificates in ", store.Location())
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	cmd.Flags().IPSliceVar(&sans.IPs, "ips", sans.IPs, "Alternative IP addresses")
	cmd.Flags().StringSliceVar(&sans.DNSNames, "domains", sans.DNSNames, "Alternative Domain names")
	return cmd
}

func merge(cn string, sans []string) []string {
	var found bool
	for _, name := range sans {
		if strings.EqualFold(name, cn) {
			found = true
			break
		}
	}
	if found {
		return sans
	}
	return append([]string{cn}, sans...)
}

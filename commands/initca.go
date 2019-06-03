package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/appscode/go/term"
	"github.com/appscode/guard/auth"
	"github.com/golang/glog"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gomodules.xyz/cert/certstore"
)

func NewCmdInitCA() *cobra.Command {
	var (
		rootDir = auth.DefaultDataDir
	)
	cmd := &cobra.Command{
		Use:               "ca",
		Short:             "Init CA",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(rootDir, "pki"))
			if err != nil {
				glog.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if store.IsExists("ca") {
				if !term.Ask(fmt.Sprintf("CA certificate found at %s. Do you want to overwrite?", store.Location()), false) {
					os.Exit(1)
				}
			}

			err = store.NewCA()
			if err != nil {
				glog.Fatalf("Failed to init ca. Reason: %v.", err)
			}
			term.Successln("Wrote ca certificates in ", store.Location())
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	return cmd
}

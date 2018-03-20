package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/appscode/go/log"
	"github.com/appscode/go/term"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var (
	rootDir = filepath.Join(homedir.HomeDir(), ".guard")
)

func NewCmdInitCA() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "ca",
		Short:             "Init CA",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			store, err := certstore.NewCertStore(afero.NewOsFs(), filepath.Join(rootDir, "pki"))
			if err != nil {
				log.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if store.IsExists("ca") {
				if !term.Ask(fmt.Sprintf("CA certificate found at %s. Do you want to overwrite?", store.Location()), false) {
					os.Exit(1)
				}
			}

			err = store.NewCA()
			if err != nil {
				log.Fatalf("Failed to init ca. Reason: %v.", err)
			}
			term.Successln("Wrote ca certificates in ", store.Location())
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	return cmd
}

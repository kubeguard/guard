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
	"fmt"
	"os"
	"path/filepath"

	"go.kubeguard.dev/guard/auth"

	"github.com/spf13/cobra"
	"gomodules.xyz/blobfs"
	"gomodules.xyz/cert/certstore"
	"gomodules.xyz/x/term"
	"k8s.io/klog/v2"
)

func NewCmdInitCA() *cobra.Command {
	rootDir := auth.DefaultDataDir
	cmd := &cobra.Command{
		Use:               "ca",
		Short:             "Init CA",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			store, err := certstore.New(blobfs.NewOsFs(), filepath.Join(rootDir, "pki"))
			if err != nil {
				klog.Fatalf("Failed to create certificate store. Reason: %v.", err)
			}
			if store.IsExists("ca") {
				if !term.Ask(fmt.Sprintf("CA certificate found at %s. Do you want to overwrite?", store.Location()), false) {
					os.Exit(1)
				}
			}

			err = store.NewCA()
			if err != nil {
				klog.Fatalf("Failed to init ca. Reason: %v.", err)
			}
			term.Successln("Wrote ca certificates in", store.Location())
		},
	}

	cmd.Flags().StringVar(&rootDir, "pki-dir", rootDir, "Path to directory where pki files are stored.")
	return cmd
}

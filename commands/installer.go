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

	"go.kubeguard.dev/guard/installer"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewCmdInstaller() *cobra.Command {
	authopts := installer.NewAuthOptions()
	authzopts := installer.NewAuthzOptions()

	cmd := &cobra.Command{
		Use:               "installer",
		Short:             "Prints Kubernetes objects for deploying guard server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			errs := authopts.Validate()
			if errs != nil {
				klog.Fatal(errs)
			}

			errs = authzopts.Validate(&authopts)
			if errs != nil {
				klog.Fatal(errs)
			}

			data, err := installer.Generate(authopts, authzopts)
			if err != nil {
				klog.Fatal(err)
			}
			fmt.Println(string(data))
		},
	}
	authopts.AddFlags(cmd.Flags())
	authzopts.AddFlags(cmd.Flags())
	return cmd
}

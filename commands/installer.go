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

	"github.com/appscode/guard/installer"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func NewCmdInstaller() *cobra.Command {
	opts := installer.New()
	cmd := &cobra.Command{
		Use:               "installer",
		Short:             "Prints Kubernetes objects for deploying guard server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			errs := opts.Validate()
			if errs != nil {
				glog.Fatal(errs)
			}

			data, err := installer.Generate(opts)
			if err != nil {
				glog.Fatal(err)
			}
			fmt.Println(string(data))
		},
	}
	opts.AddFlags(cmd.Flags())
	return cmd
}

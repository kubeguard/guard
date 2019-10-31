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
	v "github.com/appscode/go/version"
	"github.com/appscode/guard/server"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"kmodules.xyz/client-go/tools/cli"
)

func NewCmdRun() *cobra.Command {
	o := server.NewRecommendedOptions()
	srv := server.Server{
		RecommendedOptions: o,
	}
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run server",
		DisableAutoGenTag: true,
		PreRun: func(c *cobra.Command, args []string) {
			cli.SendPeriodicAnalytics(c, v.Version.Version)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if !srv.RecommendedOptions.SecureServing.UseTLS() {
				glog.Fatalln("Guard server must use SSL.")
			}
			srv.ListenAndServe()
		},
	}
	srv.AddFlags(cmd.Flags())
	return cmd
}

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
	"github.com/spf13/cobra"
	v "gomodules.xyz/x/version"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "guard [command]",
		Short:              `Guard by AppsCode - Kubernetes Authentication WebHook Server`,
		DisableAutoGenTag:  true,
		DisableFlagParsing: true,
	}
	cmd.AddCommand(NewCmdInit())
	cmd.AddCommand(NewCmdGet())
	cmd.AddCommand(NewCmdRun())
	cmd.AddCommand(NewCmdLogin())
	cmd.AddCommand(NewCmdCheck())
	cmd.AddCommand(v.NewCmdVersion())
	return cmd
}

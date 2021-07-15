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
	"strings"

	"go.kubeguard.dev/guard/auth/providers/eks"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewCmdLogin() *cobra.Command {
	var cluster, provider string

	cmd := &cobra.Command{
		Use:               "login",
		Short:             "Kubectl credential plugin",
		Long:              "Kubectl credential plugin. Visit here for more info: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			provider = strings.ToLower(provider)
			switch provider {
			case eks.OrgType:
				token, err := eks.Get(cluster)
				if err != nil {
					klog.Fatal(err)
				}
				printToken, err := eks.PrintToken(token)
				if err != nil {
					klog.Fatal(err)
				}
				fmt.Println(printToken)
				return
			case "":
				klog.Fatalln("Missing cloud provider name. Set flag -p.")
			default:
				klog.Fatalf("Unsupported cloud provider %s.", provider)
			}
		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "k", cluster, "Name of cluster")
	cmd.Flags().StringVarP(&provider, "provider", "p", provider, "Name of cloud provider")
	return cmd
}

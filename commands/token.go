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

	"github.com/appscode/guard/auth"
	"github.com/appscode/guard/auth/providers/github"
	"github.com/appscode/guard/auth/providers/gitlab"
	"github.com/appscode/guard/auth/providers/google"
	"github.com/appscode/guard/auth/providers/ldap"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

type tokenOptions struct {
	Org  string
	LDAP ldap.TokenOptions
}

func NewCmdGetToken() *cobra.Command {
	opts := tokenOptions{}

	cmd := &cobra.Command{
		Use:               "token",
		Short:             fmt.Sprintf("Get tokens for %v", auth.SupportedOrgs),
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			opts.Org = strings.ToLower(opts.Org)
			switch opts.Org {
			case github.OrgType:
				github.IssueToken()
				return
			case gitlab.OrgType:
				gitlab.IssueToken()
				return
			case google.OrgType:
				err := google.IssueToken()
				if err != nil {
					glog.Fatal(err)
				}
				return
			case ldap.OrgType:
				err := opts.LDAP.IssueToken()
				if err != nil {
					glog.Fatal("For LDAP:", err)
				}
				return
			case "":
				glog.Fatalln("Missing organization name. Set flag -o Google|Github.")
			default:
				glog.Fatalf("Unknown organization %s.", opts.Org)
			}
		},
	}

	cmd.Flags().StringVarP(&opts.Org, "organization", "o", opts.Org, fmt.Sprintf("Name of Organization (%v).", auth.SupportedOrgs))
	opts.LDAP.AddFlags(cmd.Flags())
	return cmd
}

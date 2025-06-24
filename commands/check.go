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
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"go.kubeguard.dev/guard/auth"
	"go.kubeguard.dev/guard/auth/providers/azure"
	"go.kubeguard.dev/guard/auth/providers/github"
	"go.kubeguard.dev/guard/auth/providers/gitlab"
	"go.kubeguard.dev/guard/auth/providers/google"
	"go.kubeguard.dev/guard/auth/providers/ldap"
	"go.kubeguard.dev/guard/server"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCmdCheck() *cobra.Command {
	var (
		opts         = server.NewAuthRecommendedOptions()
		githubOrg    string
		googleDomain string
	)
	cmd := &cobra.Command{
		Use:   "check PROVIDER",
		Short: "Checks a token from STDIN and prints the obtained user information.",
		Long: `Checks a token from STDIN and prints the obtained user information. The check
command accepts the same flags as the run command. This way you can verify that
your configuration would produce the expected result for a certain token.

For the Github and Google provider you additionally have to configure the flag
--github.org resp. --google.domain. In the running server these values are passed
as the common name (CN) of the client certificate.`,
		Args:              cobra.MinimumNArgs(1),
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]

			var (
				authenticator auth.Interface
				err           error
			)
			switch strings.ToLower(provider) {
			case github.OrgType:
				if githubOrg == "" {
					return fmt.Errorf("github organization not configured. set it with --github.org")
				}
				authenticator = github.New(opts.Github, githubOrg)
			case google.OrgType:
				if googleDomain == "" {
					return fmt.Errorf("google domain not configured. set it with --google.domain")
				}
				authenticator, err = google.New(cmd.Context(), opts.Google, googleDomain)
			case gitlab.OrgType:
				authenticator = gitlab.New(opts.Gitlab)
			case azure.OrgType:
				authenticator, err = azure.New(cmd.Context(), opts.Azure)
			case ldap.OrgType:
				authenticator = ldap.New(opts.LDAP)
			default:
				return errors.Errorf("unknown provider '%s'", provider)
			}
			if err != nil {
				return err
			}

			token, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("failed to read token from stdin")
			}

			userinfo, err := authenticator.Check(cmd.Context(), string(token))
			if err != nil {
				return err
			}

			out, err := json.MarshalIndent(userinfo, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(out))
			return nil
		},
	}

	cmd.Flags().StringVar(&githubOrg, "github.org", githubOrg, "Github organization. With the running guard server this value is passed as common name (CN) of the client certificate")
	cmd.Flags().StringVar(&googleDomain, "google.domain", googleDomain, "Google domain. With the running guard server this value is passed as common name (CN) of the client certificate")
	opts.AddFlags(cmd.Flags())
	return cmd
}

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

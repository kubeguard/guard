package cmds

import (
	"fmt"
	"strings"

	"github.com/appscode/go/log"
	"github.com/appscode/guard/appscode"
	"github.com/appscode/guard/github"
	"github.com/appscode/guard/gitlab"
	"github.com/appscode/guard/google"
	"github.com/appscode/guard/server"
	"github.com/spf13/cobra"
)

func NewCmdGetToken() *cobra.Command {
	var org string
	cmd := &cobra.Command{
		Use:               "token",
		Short:             fmt.Sprintf("Get tokens for %v", server.SupportedOrgPrintForm()),
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			org = strings.ToLower(org)
			switch org {
			case github.OrgType:
				github.IssueToken()
			case gitlab.OrgType:
				gitlab.IssueToken()
			case google.OrgType:
				google.IssueToken()
			case appscode.OrgType:
				appscode.IssueToken()
			case "":
				log.Fatalln("Missing organization name. Set flag -o Google|Github.")
			default:
				log.Fatalf("Unknown organization %s.", org)
			}
		},
	}

	cmd.Flags().StringVarP(&org, "organization", "o", org, fmt.Sprintf("Name of Organization (%v).", server.SupportedOrgPrintForm()))
	return cmd
}

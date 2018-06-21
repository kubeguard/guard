package commands

import (
	"fmt"
	"strings"

	"github.com/appscode/guard/auth"
	"github.com/appscode/guard/auth/providers/aws"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func NewCmdGetClusterToken() *cobra.Command {
	var cluster, provider string

	cmd := &cobra.Command{
		Use:               "cluster-token",
		Short:             fmt.Sprintf("Get tokens for %v", auth.SupportedOrgs),
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			provider = strings.ToLower(provider)
			switch provider {
			case aws.OrgType:
				token, err := aws.Get(cluster)
				if err != nil {
					glog.Fatal(err)
				}
				printToken, err := aws.PrintToken(token)
				if err != nil {
					glog.Fatal(err)
				}
				fmt.Println(printToken)
				return
			case "":
				glog.Fatalln("Missing cloud provider name. Set flag -p eks.")
			default:
				glog.Fatalf("Unknown cloud provider %s.", provider)
			}

		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "k", cluster, fmt.Sprintf("Name of Cluster"))
	cmd.Flags().StringVarP(&provider, "provider", "p", provider, fmt.Sprintf("Name of Cloud provider"))
	return cmd
}

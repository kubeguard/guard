package commands

import (
	"fmt"
	"strings"

	"github.com/appscode/guard/auth/providers/eks"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
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
					glog.Fatal(err)
				}
				printToken, err := eks.PrintToken(token)
				if err != nil {
					glog.Fatal(err)
				}
				fmt.Println(printToken)
				return
			case "":
				glog.Fatalln("Missing cloud provider name. Set flag -p.")
			default:
				glog.Fatalf("Unsupported cloud provider %s.", provider)
			}
		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "k", cluster, fmt.Sprintf("Name of cluster"))
	cmd.Flags().StringVarP(&provider, "provider", "p", provider, fmt.Sprintf("Name of cloud provider"))
	return cmd
}

package commands

import (
	"flag"
	"log"

	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/tools/cli"
	"github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "guard [command]",
		Short:              `Guard by AppsCode - Kubernetes Authentication WebHook Server`,
		DisableAutoGenTag:  true,
		DisableFlagParsing: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			cli.SendAnalytics(c, v.Version.Version)
		},
	}
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	cmd.PersistentFlags().BoolVar(&cli.EnableAnalytics, "analytics", cli.EnableAnalytics, "Send analytical events to Google Guard")

	cmd.AddCommand(NewCmdInit())
	cmd.AddCommand(NewCmdGet())
	cmd.AddCommand(NewCmdRun())
	cmd.AddCommand(NewCmdLogin())
	cmd.AddCommand(v.NewCmdVersion())
	return cmd
}

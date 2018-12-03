package commands

import (
	"flag"

	"github.com/appscode/go/flags"
	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/tools/cli"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "guard [command]",
		Short:              `Guard by AppsCode - Kubernetes Authentication WebHook Server`,
		DisableAutoGenTag:  true,
		DisableFlagParsing: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			flags.DumpAll(c.Flags())
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

package commands

import (
	"flag"
	"log"
	"strings"

	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/tools/analytics"
	"github.com/jpillora/go-ogle-analytics"
	"github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	gaTrackingCode = "UA-62096468-20"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func NewRootCmd() *cobra.Command {
	var (
		enableAnalytics = true
	)
	cmd := &cobra.Command{
		Use:                "guard [command]",
		Short:              `Guard by AppsCode - Kubernetes Authentication WebHook Server`,
		DisableAutoGenTag:  true,
		DisableFlagParsing: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			if enableAnalytics && gaTrackingCode != "" {
				if client, err := ga.NewClient(gaTrackingCode); err == nil {
					client.ClientID(analytics.ClientID())
					parts := strings.Split(c.CommandPath(), " ")
					client.Send(ga.NewEvent(parts[0], strings.Join(parts[1:], "/")).Label(v.Version.Version))
				}
			}
		},
	}
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	cmd.PersistentFlags().BoolVar(&enableAnalytics, "analytics", enableAnalytics, "Send analytical events to Google Guard")

	cmd.AddCommand(NewCmdInit())
	cmd.AddCommand(NewCmdGet())
	cmd.AddCommand(NewCmdRun())
	cmd.AddCommand(NewCmdLogin())
	cmd.AddCommand(v.NewCmdVersion())
	return cmd
}

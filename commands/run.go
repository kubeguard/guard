package commands

import (
	v "github.com/appscode/go/version"
	"github.com/appscode/guard/server"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"kmodules.xyz/client-go/tools/cli"
)

func NewCmdRun() *cobra.Command {
	o := server.NewRecommendedOptions()
	srv := server.Server{
		RecommendedOptions: o,
	}
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run server",
		DisableAutoGenTag: true,
		PreRun: func(c *cobra.Command, args []string) {
			cli.SendPeriodicAnalytics(c, v.Version.Version)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if !srv.RecommendedOptions.SecureServing.UseTLS() {
				glog.Fatalln("Guard server must use SSL.")
			}
			srv.ListenAndServe()
		},
	}
	srv.AddFlags(cmd.Flags())
	return cmd
}

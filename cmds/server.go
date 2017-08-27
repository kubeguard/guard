package cmds

import (
	"github.com/appscode/kad/analytics"
	"github.com/appscode/kad/server"
	"github.com/spf13/cobra"
)

func NewCmdServer(version string) *cobra.Command {
	srv := hostfacts.Server{
		WebAddress:      ":9844",
		OpsAddress:      ":56790",
		EnableAnalytics: true,
	}
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run server",
		DisableAutoGenTag: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if srv.EnableAnalytics {
				analytics.Enable()
			}
			analytics.SendEvent("kad", "started", version)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			analytics.SendEvent("kad", "stopped", version)
		},
		Run: func(cmd *cobra.Command, args []string) {
			srv.ListenAndServe()
		},
	}

	cmd.Flags().StringVar(&srv.WebAddress, "web-address", srv.WebAddress, "Http server address")
	cmd.Flags().StringVar(&srv.CACertFile, "caCertFile", srv.CACertFile, "File containing CA certificate")
	cmd.Flags().StringVar(&srv.CertFile, "certFile", srv.CertFile, "File container server TLS certificate")
	cmd.Flags().StringVar(&srv.KeyFile, "keyFile", srv.KeyFile, "File containing server TLS private key")

	cmd.Flags().StringVar(&srv.OpsAddress, "ops-addr", srv.OpsAddress, "Address to listen on for web interface and telemetry.")
	cmd.Flags().BoolVar(&srv.EnableAnalytics, "kad", srv.EnableAnalytics, "Send analytical events to Google Kad")
	return cmd
}

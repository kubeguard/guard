package cmds

import (
	"github.com/appscode/kad/lib"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	srv := lib.Server{
		WebAddress: ":9844",
		OpsAddress: ":56790",
	}
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if !srv.UseTLS() {
				log.Fatalln("Kad server must use SSL.")
			}
			srv.ListenAndServe()
		},
	}

	cmd.Flags().StringVar(&srv.WebAddress, "web-address", srv.WebAddress, "Http server address")
	cmd.Flags().StringVar(&srv.CACertFile, "ca-cert-file", srv.CACertFile, "File containing CA certificate")
	cmd.Flags().StringVar(&srv.CertFile, "cert-file", srv.CertFile, "File container server TLS certificate")
	cmd.Flags().StringVar(&srv.KeyFile, "key-file", srv.KeyFile, "File containing server TLS private key")

	cmd.Flags().StringVar(&srv.OpsAddress, "ops-addr", srv.OpsAddress, "Address to listen on for web interface and telemetry.")
	return cmd
}

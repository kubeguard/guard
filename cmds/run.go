package cmds

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/guard/lib"
	"github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/analytics"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	webPort = 9844
	opsPort = 56790
)

func NewCmdRun() *cobra.Command {
	srv := lib.Server{
		WebAddress: fmt.Sprintf(":%d", webPort),
		OpsAddress: fmt.Sprintf(":%d", opsPort),
	}
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run server",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if !srv.UseTLS() {
				log.Fatalln("Guard server must use SSL.")
			}

			if meta.PossiblyInCluster() {
				config, err := rest.InClusterConfig()
				if err != nil {
					log.Fatalln(err)
				}
				kubeClient := kubernetes.NewForConfigOrDie(config)
				sendAnalytics(cmd, analytics.ClientID(kubeClient.CoreV1().Nodes()))
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

package cmds

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/guard/lib"
	"github.com/spf13/cobra"
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
			srv.ListenAndServe()
		},
	}
	srv.AddFlags(cmd.Flags())
	return cmd
}

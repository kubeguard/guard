package cmds

import (
	"github.com/spf13/cobra"
)

func NewCmdInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "init",
		Short:             `Init PKI`,
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(NewCmdInitCA())
	cmd.AddCommand(NewCmdInitServer())
	cmd.AddCommand(NewCmdInitClient())
	return cmd
}

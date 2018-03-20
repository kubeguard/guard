package main

import (
	"os"

	logs "github.com/appscode/go/log/golog"
	"github.com/appscode/guard/commands"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/vjeantet/ldapserver"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := commands.NewRootCmdCLI(Version).Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

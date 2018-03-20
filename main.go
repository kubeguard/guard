package main

import (
	"os"

	logs "github.com/appscode/go/log/golog"
	_ "github.com/appscode/guard/auth/providers"
	"github.com/appscode/guard/commands"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/vjeantet/ldapserver"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := commands.NewRootCmd(Version).Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

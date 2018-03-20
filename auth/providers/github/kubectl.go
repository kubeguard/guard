package github

import (
	"github.com/appscode/go/term"
	"github.com/skratchdot/open-golang/open"
)

func IssueToken() {
	codeURurl := "https://github.com/settings/tokens/new"
	term.Infoln("Github url for personal access tokens:", codeURurl)
	open.Start(codeURurl)
}

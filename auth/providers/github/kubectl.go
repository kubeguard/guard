package github

import (
	"github.com/appscode/go/term"
	"github.com/skratchdot/open-golang/open"
)

func IssueToken() {
	codeURurl := "https://github.com/settings/tokens/new"
	term.Infoln("Github url for personal access tokens:", codeURurl)
	term.Infoln("Select only read:org permissions for the token")
	term.Infoln("After the token is created, run:")
	term.Infoln("    kubectl config set-credentials <user_name> --token=<token>")
	open.Start(codeURurl)
}

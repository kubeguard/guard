package gitlab

import (
	"github.com/appscode/go/term"
	"github.com/skratchdot/open-golang/open"
)

func IssueToken() {
	codeURurl := "https://gitlab.com/profile/personal_access_tokens"
	term.Infoln("Gitlab url for personal access tokens:", codeURurl)
	open.Start(codeURurl)
}

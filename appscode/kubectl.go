package appscode

import (
	"fmt"
	"strings"

	"github.com/appscode/go/term"
	"github.com/howeyc/gopass"
	"github.com/skratchdot/open-golang/open"
)

func IssueToken() error {
	teamId := term.Read("Team Id:")
	endpoint := fmt.Sprintf("https://%v.appscode.io", teamId)
	err := open.Start(strings.Join([]string{endpoint, "conduit", "login"}, "/"))

	term.Print("Paste the token here: ")
	tokenBytes, err := gopass.GetPasswdMasked()
	if err != nil {
		term.Fatalln("Failed to retrieve token", err)
	}

	token := string(tokenBytes)
	client := &ConduitClient{
		Url:   strings.Join([]string{endpoint, "api", "user.whoami"}, "/"),
		Token: token,
	}
	result := &WhoAmIResponse{}
	err = client.Call().Into(result)
	if err != nil {
		term.Fatalln("Failed to validate token", err)
	}
	if result.ErrorCode != nil {
		term.Fatalln("Failed to validate token")
	}
	term.Successln("Token successfully generated")

	return err
}

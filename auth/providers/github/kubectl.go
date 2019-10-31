/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

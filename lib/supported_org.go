package lib

import "strings"

func GetSupportedOrg() []string {
	return []string{
		"Github",
		"Gitlab",
		"Google",
	}
}

// output form : Github/Google/Gitlab
func SupportedOrgPrintForm() string {
	return strings.Join(GetSupportedOrg(), "/")
}

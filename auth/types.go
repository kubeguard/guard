package auth

import "strings"

type orgs []string

var SupportedOrgs orgs

func (o orgs) Has(name string) bool {
	name = strings.TrimSpace(strings.ToLower(name))
	for _, org := range o {
		if org == name {
			return true
		}
	}
	return false
}

func (o orgs) String() string {
	return strings.Join(o, "/")
}

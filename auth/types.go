package auth

import (
	"path/filepath"
	"strings"

	"k8s.io/client-go/util/homedir"
)

var DefaultPKIDir = filepath.Join(homedir.HomeDir(), ".guard")

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
	names := make([]string, len(o))
	for i, org := range o {
		names[i] = strings.Title(org)
	}
	return strings.Join(names, "/")
}

package auth

import (
	"path/filepath"
	"sort"
	"strings"

	authv1 "k8s.io/api/authentication/v1"
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
	sort.Strings(names)
	return strings.Join(names, "/")
}

type Interface interface {
	UID() string
	Check(token string) (*authv1.UserInfo, error)
}

package auth

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang/glog"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/client-go/util/homedir"
)

var DefaultDataDir = func() string {
	if v, ok := os.LookupEnv("GUARD_DATA_DIR"); ok {
		glog.Infof("Using data dir %s found in GUARD_DATA_DIR env variable", v)
		return v
	}
	return filepath.Join(homedir.HomeDir(), ".guard")
}()

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

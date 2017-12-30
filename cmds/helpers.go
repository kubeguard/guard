package cmds

import "k8s.io/client-go/util/cert"

func filename(cfg cert.Config) string {
	if len(cfg.Organization) == 0 {
		return cfg.CommonName
	}
	return cfg.CommonName + "@" + cfg.Organization[0]
}

package commands

import "k8s.io/client-go/util/cert"

func filename(cfg cert.Config) string {
	if len(cfg.Organization) == 0 {
		return cfg.AltNames.DNSNames[0]
	}
	return cfg.AltNames.DNSNames[0] + "@" + cfg.Organization[0]
}

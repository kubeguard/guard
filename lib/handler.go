package lib

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	auth "k8s.io/api/authentication/v1"
)

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	crt := req.TLS.PeerCertificates[0]
	if len(crt.Subject.Organization) == 0 {
		Write(w, Error("Client certificate is missing organization"), http.StatusBadRequest)
		return
	}
	org := crt.Subject.Organization[0]
	log.Infof("Received token review request for %s@%s", crt.Subject.CommonName, org)

	data := auth.TokenReview{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		Write(w, Error("Failed to parse request. Reason: "+err.Error()), http.StatusBadRequest)
		return
	}

	if s.TokenAuthFile != "" {
		resp, code := s.checkTokenAuth(data.Spec.Token)
		if resp.Status.Authenticated {
			Write(w, resp, code)
			return
		}
	}

	switch strings.ToLower(org) {
	case "github":
		client := NewGithubClient(crt.Subject.CommonName, data.Spec.Token)
		resp, code := client.checkGithub()
		Write(w, resp, code)
		return
	case "google":
		resp, code := s.checkGoogle(crt.Subject.CommonName, data.Spec.Token)
		Write(w, resp, code)
		return
	case "appscode":
		resp, code := checkAppscode(crt.Subject.CommonName, data.Spec.Token)
		Write(w, resp, code)
		return
	case "gitlab":
		resp, code := checkGitLab(data.Spec.Token)
		Write(w, resp, code)
		return
	case "azure":
		if s.Azure.ClientID == "" || s.Azure.ClientSecret == "" || s.Azure.TenantID == "" {
			Write(w, Error("Missing azure client-id or client-secret or tenant-id"), http.StatusBadRequest)
		}
		resp, code := s.checkAzure(data.Spec.Token)
		Write(w, resp, code)
		return
	case "ldap":
		resp, code := s.checkLDAP(data.Spec.Token)
		Write(w, resp, code)
		return
	}
	Write(w, Error("Client is using unknown organization "+org), http.StatusBadRequest)
	return
}

package server

import (
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	"github.com/appscode/guard/auth/providers/appscode"
	"github.com/appscode/guard/auth/providers/azure"
	"github.com/appscode/guard/auth/providers/github"
	"github.com/appscode/guard/auth/providers/gitlab"
	"github.com/appscode/guard/auth/providers/google"
	"github.com/appscode/guard/auth/providers/ldap"
	"github.com/pkg/errors"
	auth "k8s.io/api/authentication/v1"
)

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
		write(w, nil, WithCode(errors.New("Missing client certificate"), http.StatusBadRequest))
		return
	}
	crt := req.TLS.PeerCertificates[0]
	if len(crt.Subject.Organization) == 0 {
		write(w, nil, WithCode(errors.New("Client certificate is missing organization"), http.StatusBadRequest))
		return
	}
	org := crt.Subject.Organization[0]
	log.Infof("Received token review request for %s/%s", org, crt.Subject.CommonName)

	data := auth.TokenReview{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		write(w, nil, WithCode(errors.Wrap(err, "Failed to parse request"), http.StatusBadRequest))
		return
	}

	if s.TokenAuthenticator != nil {
		resp, err := s.TokenAuthenticator.Check(data.Spec.Token)
		if err == nil {
			write(w, resp, err)
			return
		}
	}

	switch strings.ToLower(org) {
	case github.OrgType:
		client := github.New(crt.Subject.CommonName, data.Spec.Token)
		resp, err := client.Check()
		write(w, resp, err)
		return
	case google.OrgType:
		client, err := google.New(s.RecommendedOptions.Google, crt.Subject.CommonName)
		if err != nil {
			write(w, nil, err)
			return
		}
		resp, err := client.Check(crt.Subject.CommonName, data.Spec.Token)
		write(w, resp, err)
		return
	case appscode.OrgType:
		resp, err := appscode.Check(crt.Subject.CommonName, data.Spec.Token)
		write(w, resp, err)
		return
	case gitlab.OrgType:
		client := gitlab.New(data.Spec.Token)
		resp, err := client.Check()
		write(w, resp, err)
		return
	case azure.OrgType:
		if s.RecommendedOptions.Azure.ClientID == "" || s.RecommendedOptions.Azure.ClientSecret == "" || s.RecommendedOptions.Azure.TenantID == "" {
			write(w, nil, errors.New("Missing azure client-id or client-secret or tenant-id"))
			return
		}
		client, err := azure.New(s.RecommendedOptions.Azure)
		if err != nil {
			write(w, nil, err)
			return
		}
		resp, err := client.Check(data.Spec.Token)
		write(w, resp, err)
		return
	case ldap.OrgType:
		client, err := ldap.New(s.RecommendedOptions.LDAP)
		if err != nil {
			write(w, nil, err)
			return
		}
		resp, code := client.Check(data.Spec.Token)
		write(w, resp, code)
		return
	}
	write(w, nil, WithCode(errors.Errorf("Client is using unknown organization %s", org), http.StatusBadRequest))
	return
}

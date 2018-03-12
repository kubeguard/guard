package server

import (
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	"github.com/appscode/guard/appscode"
	"github.com/appscode/guard/azure"
	"github.com/appscode/guard/github"
	"github.com/appscode/guard/gitlab"
	"github.com/appscode/guard/google"
	"github.com/appscode/guard/ldap"
	"github.com/pkg/errors"
	auth "k8s.io/api/authentication/v1"
)

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	crt := req.TLS.PeerCertificates[0]
	if len(crt.Subject.Organization) == 0 {
		Write(w, nil, errors.New("Client certificate is missing organization"))
		return
	}
	org := crt.Subject.Organization[0]
	log.Infof("Received token review request for %s/%s", org, crt.Subject.CommonName)

	data := auth.TokenReview{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		Write(w, nil, errors.Wrap(err, "Failed to parse request"))
		return
	}

	if s.TokenAuthenticator != nil {
		resp, err := s.TokenAuthenticator.Check(data.Spec.Token)
		if err == nil {
			Write(w, resp, err)
			return
		}
	}

	switch strings.ToLower(org) {
	case github.OrgType:
		client := github.New(crt.Subject.CommonName, data.Spec.Token)
		resp, err := client.Check()
		Write(w, resp, err)
		return
	case google.OrgType:
		client, err := google.New(s.RecommendedOptions.Google, crt.Subject.CommonName)
		if err != nil {
			Write(w, nil, err)
			return
		}
		resp, err := client.Check(crt.Subject.CommonName, data.Spec.Token)
		Write(w, resp, err)
		return
	case appscode.OrgType:
		resp, err := appscode.Check(crt.Subject.CommonName, data.Spec.Token)
		Write(w, resp, err)
		return
	case gitlab.OrgType:
		client := gitlab.New(data.Spec.Token)
		resp, err := client.Check()
		Write(w, resp, err)
		return
	case azure.OrgType:
		if s.RecommendedOptions.Azure.ClientID == "" || s.RecommendedOptions.Azure.ClientSecret == "" || s.RecommendedOptions.Azure.TenantID == "" {
			Write(w, nil, errors.New("Missing azure client-id or client-secret or tenant-id"))
			return
		}
		client, err := azure.New(s.RecommendedOptions.Azure)
		if err != nil {
			Write(w, nil, err)
			return
		}
		resp, err := client.Check(data.Spec.Token)
		Write(w, resp, err)
		return
	case ldap.OrgType:
		client := ldap.New(s.RecommendedOptions.LDAP)
		resp, code := client.Check(data.Spec.Token)
		Write(w, resp, code)
		return
	}
	Write(w, nil, errors.Errorf("Client is using unknown organization %s", org))
	return
}

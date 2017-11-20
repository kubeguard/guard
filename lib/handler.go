package lib

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	auth "k8s.io/api/authentication/v1beta1"
)

func Authenticate(w http.ResponseWriter, req *http.Request) {
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

	switch strings.ToLower(org) {
	case "github":
		resp, code := checkGithub(crt.Subject.CommonName, data.Spec.Token)
		Write(w, resp, code)
		return
	case "google":
		resp, code := checkGoogle(crt.Subject.CommonName, data.Spec.Token)
		Write(w, resp, code)
		return
	case "appscode":
		resp, code := checkAppscode(crt.Subject.CommonName, data.Spec.Token)
		Write(w, resp, code)
		return
	}
	Write(w, Error("Client is using unknown organization "+org), http.StatusBadRequest)
	return
}

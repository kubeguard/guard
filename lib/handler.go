package lib

import (
	"encoding/json"
	"net/http"
	"strings"

	auth "k8s.io/client-go/pkg/apis/authentication/v1beta1"
)

func Authenticate(w http.ResponseWriter, req *http.Request) {
	crt := req.TLS.PeerCertificates[0]

	if len(crt.Subject.Organization) == 0 {
		Error(w, "Client certificate is missing organization", http.StatusBadRequest)
		return
	}
	org := crt.Subject.Organization[0]

	data := auth.TokenReview{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		Error(w, "Failed to parse request. Reason: "+err.Error(), http.StatusBadRequest)
		return
	}
	switch strings.ToLower(org) {
	case "github":
		checkGithub(w, crt.Subject.CommonName, data.Spec.Token)
		return
	case "google":
		checkGoogle(w, crt.Subject.CommonName, data.Spec.Token)
		return
	}
	Error(w, "Client is using unknown organization "+org, http.StatusBadRequest)
	return
}

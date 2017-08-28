package lib

import (
	"encoding/json"
	"net/http"
	"strings"
)

func Authenticate(w http.ResponseWriter, req *http.Request) {
	crt := req.TLS.PeerCertificates[0]

	if len(crt.Subject.Organization) == 0 {
		Error(w, "Client certificate is missing organization", http.StatusBadRequest)
		return
	}
	org := crt.Subject.Organization[0]

	data := NewTokenReview()
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

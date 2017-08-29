package lib

import (
	"encoding/json"
	"net/http"
	"strings"

	auth "k8s.io/client-go/pkg/apis/authentication/v1beta1"
	"fmt"
	"github.com/tamalsaha/go-oneliners"
)

func Authenticate(w http.ResponseWriter, req *http.Request) {
	crt := req.TLS.PeerCertificates[0]

	if len(crt.Subject.Organization) == 0 {
		Write(w, Error("Client certificate is missing organization"), http.StatusBadRequest)
		return
	}
	org := crt.Subject.Organization[0]

	data := auth.TokenReview{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		Write(w, Error("Failed to parse request. Reason: "+err.Error()), http.StatusBadRequest)
		return
	}

	pb, _ := json.Marshal(data)
	fmt.Println(string(pb))

	switch strings.ToLower(org) {
	case "github":
		oneliners.FILE()
		resp, code := checkGithub(crt.Subject.CommonName, data.Spec.Token)
		Write(w, resp, code)
		return
	case "google":
		resp, code := checkGoogle(crt.Subject.CommonName, data.Spec.Token)
		Write(w, resp, code)
		return
	}
	Write(w, Error("Client is using unknown organization "+org), http.StatusBadRequest)
	return
}

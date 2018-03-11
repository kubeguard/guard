package server

import (
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	auth "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Write replies to the request with the specified TokenReview object and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
func Write(w http.ResponseWriter, info *auth.UserInfo, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-content-type-options", "nosniff")

	resp := auth.TokenReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: auth.SchemeGroupVersion.String(),
			Kind:       "TokenReview",
		},
	}

	if err != nil {
		printStackTrace(err)
		w.WriteHeader(http.StatusUnauthorized)
		resp.Status = auth.TokenReviewStatus{
			Authenticated: false,
			Error:         err.Error(),
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	resp.Status = auth.TokenReviewStatus{
		Authenticated: true,
		User:          *info,
	}
	json.NewEncoder(w).Encode(resp)
}

func printStackTrace(e2 error) {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	err, ok := errors.Cause(e2).(stackTracer)
	if !ok {
		panic("oops, err does not implement stackTracer")
	}

	st := err.StackTrace()
	log.Errorf("%s\nStacktrace: %+v", e2.Error(), st) // top two frames
}

func GetSupportedOrg() []string {
	return []string{
		"Github",
		"Gitlab",
		"Google",
	}
}

// output form : Github/Google/Gitlab
func SupportedOrgPrintForm() string {
	return strings.Join(GetSupportedOrg(), "/")
}

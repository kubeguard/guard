package lib

import (
	"encoding/json"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	auth "k8s.io/client-go/pkg/apis/authentication/v1beta1"
)

const apiVersion = "authentication.k8s.io/v1beta1"

// Write replies to the request with the specified TokenReview object and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
func Write(w http.ResponseWriter, data auth.TokenReview, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-content-type-options", "nosniff")
	w.WriteHeader(code)
	data.TypeMeta = metav1.TypeMeta{
		APIVersion: apiVersion,
		Kind:       "TokenReview",
	}
	data.Status.Authenticated = true
	json.NewEncoder(w).Encode(data)
}

// Error returns a `TokenReview` response with the specified error message.
func Error(error string) auth.TokenReview {
	return auth.TokenReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       "TokenReview",
		},
		Status: auth.TokenReviewStatus{
			Authenticated: false,
			Error:         error,
		},
	}
}

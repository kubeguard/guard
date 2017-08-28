package lib

import (
	"encoding/json"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	auth "k8s.io/client-go/pkg/apis/authentication/v1beta1"
)

func NewTokenReview() auth.TokenReview {
	return auth.TokenReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "authentication.k8s.io/v1beta1",
			Kind:       "TokenReview",
		},
	}
}

func Write(w http.ResponseWriter, data auth.TokenReview) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-content-type-options", "nosniff")
	w.WriteHeader(http.StatusOK)
	data.Status.Authenticated = true
	json.NewEncoder(w).Encode(data)
}

// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
func Error(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-content-type-options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(auth.TokenReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "authentication.k8s.io/v1beta1",
			Kind:       "TokenReview",
		},
		Status: auth.TokenReviewStatus{
			Authenticated: false,
			Error:         error,
		},
	})
}

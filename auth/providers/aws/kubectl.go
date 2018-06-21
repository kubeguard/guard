package aws

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	auth "k8s.io/client-go/pkg/apis/clientauthentication/v1alpha1"
)

func PrintToken(token string) (string, error) {
	execInput := &auth.ExecCredential{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "client.authentication.k8s.io/v1alpha1",
			Kind:       "ExecCredential",
		},
		Status: &auth.ExecCredentialStatus{
			Token: token,
		},
	}
	ret, err := json.Marshal(execInput)
	return string(ret), err
}

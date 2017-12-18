package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeCredential = "cred"
	ResourceKindCredential = "Credential"
	ResourceNameCredential = "credential"
	ResourceTypeCredential = "credentials"

	ResourceProviderCredential = "provider"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Credential struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              CredentialSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type CredentialSpec struct {
	Provider string            `json:"provider" protobuf:"bytes,1,opt,name=provider"`
	Data     map[string]string `json:"data" protobuf:"bytes,2,rep,name=data"`
}

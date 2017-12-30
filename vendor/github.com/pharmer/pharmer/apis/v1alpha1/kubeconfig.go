package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Config holds the information needed to build connect to remote kubernetes clusters as a given user
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type KubeConfig struct {
	metav1.TypeMeta `json:",inline,omitempty,omitempty"`
	// Preferences holds general information to be use for cli interactions
	Preferences Preferences `json:"preferences" protobuf:"bytes,1,opt,name=preferences"`
	// Clusters is a map of referencable names to cluster configs
	Cluster NamedCluster `json:"cluster" protobuf:"bytes,2,opt,name=cluster"`
	// AuthInfos is a map of referencable names to user configs
	AuthInfo NamedAuthInfo `json:"user" protobuf:"bytes,3,opt,name=user"`
	// Contexts is a map of referencable names to context configs
	Context NamedContext `json:"context" protobuf:"bytes,4,opt,name=context"`
}

type Preferences struct {
	// +optional
	Colors bool `json:"colors,omitempty" protobuf:"varint,1,opt,name=colors"`
}

// NamedCluster relates nicknames to cluster information
type NamedCluster struct {
	// Name is the nickname for this Cluster
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Server is the address of the kubernetes cluster (https://hostname:port).
	Server string `json:"server" protobuf:"bytes,2,opt,name=server"`
	// CertificateAuthorityData contains PEM-encoded certificate authority certificates. Overrides CertificateAuthorityData
	// +optional
	CertificateAuthorityData []byte `json:"certificateAuthorityData,omitempty" protobuf:"bytes,3,opt,name=certificateAuthorityData"`
}

// NamedContext relates nicknames to context information
type NamedContext struct {
	// Name is the nickname for this Context
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Cluster is the name of the cluster for this context
	Cluster string `json:"cluster" protobuf:"bytes,2,opt,name=cluster"`
	// AuthInfo is the name of the authInfo for this context
	AuthInfo string `json:"user" protobuf:"bytes,3,opt,name=user"`
}

// NamedAuthInfo relates nicknames to auth information
type NamedAuthInfo struct {
	// Name is the nickname for this AuthInfo
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// ClientCertificateData contains PEM-encoded data from a client cert file for TLS.
	// +optional
	ClientCertificateData []byte `json:"clientCertificateData,omitempty" protobuf:"bytes,2,opt,name=clientCertificateData"`
	// ClientKeyData contains PEM-encoded data from a client key file for TLS.
	// +optional
	ClientKeyData []byte `json:"clientKeyData,omitempty" protobuf:"bytes,3,opt,name=clientKeyData"`
	// Token is the bearer token for authentication to the kubernetes cluster.
	// +optional
	Token string `json:"token,omitempty" protobuf:"bytes,4,opt,name=token"`
}

func Convert_KubeConfig_To_Config(in *KubeConfig) *clientcmdapi.Config {
	return &clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: clientcmdapi.SchemeGroupVersion.String(),
		Preferences: clientcmdapi.Preferences{
			Colors: in.Preferences.Colors,
		},
		Clusters: map[string]*clientcmdapi.Cluster{
			in.Cluster.Name: {
				Server: in.Cluster.Server,
				CertificateAuthorityData: append([]byte(nil), in.Cluster.CertificateAuthorityData...),
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			in.AuthInfo.Name: {
				Token: in.AuthInfo.Token,
				ClientCertificateData: append([]byte(nil), in.AuthInfo.ClientCertificateData...),
				ClientKeyData:         append([]byte(nil), in.AuthInfo.ClientKeyData...),
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			in.Context.Name: {
				Cluster:  in.Context.Cluster,
				AuthInfo: in.Context.AuthInfo,
			},
		},
		CurrentContext: in.Context.Name,
	}
}

func NewRestConfig(in *KubeConfig) *rest.Config {
	out := &rest.Config{
		Host: in.Cluster.Server,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: append([]byte(nil), in.Cluster.CertificateAuthorityData...),
		},
	}
	if in.AuthInfo.Token == "" {
		out.TLSClientConfig.CertData = append([]byte(nil), in.AuthInfo.ClientCertificateData...)
		out.TLSClientConfig.KeyData = append([]byte(nil), in.AuthInfo.ClientKeyData...)
	} else {
		out.BearerToken = in.AuthInfo.Token
	}
	return out
}

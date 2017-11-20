package v1alpha1

type SSHConfig struct {
	PrivateKey []byte `json:"privateKey,omitempty" protobuf:"bytes,1,opt,name=privateKey"`
	HostIP     string `json:"hostIP,omitempty" protobuf:"bytes,2,opt,name=hostIP"`
	HostPort   int32  `json:"hostPort,omitempty" protobuf:"varint,3,opt,name=hostPort"`
	User       string `json:"user,omitempty" protobuf:"bytes,4,opt,name=user"`
}

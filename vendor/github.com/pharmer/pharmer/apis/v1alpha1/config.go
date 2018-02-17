package v1alpha1

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LocalSpec struct {
	Path string `json:"path,omitempty" protobuf:"bytes,1,opt,name=path"`
}

type S3Spec struct {
	Endpoint string `json:"endpoint,omitempty" protobuf:"bytes,1,opt,name=endpoint"`
	Bucket   string `json:"bucket,omiempty" protobuf:"bytes,2,opt,name=bucket"`
	Prefix   string `json:"prefix,omitempty" protobuf:"bytes,3,opt,name=prefix"`
}

type GCSSpec struct {
	Bucket string `json:"bucket,omiempty" protobuf:"bytes,1,opt,name=bucket"`
	Prefix string `json:"prefix,omitempty" protobuf:"bytes,2,opt,name=prefix"`
}

type AzureStorageSpec struct {
	Container string `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	Prefix    string `json:"prefix,omitempty" protobuf:"bytes,2,opt,name=prefix"`
}

type SwiftSpec struct {
	Container string `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	Prefix    string `json:"prefix,omitempty" protobuf:"bytes,2,opt,name=prefix"`
}

type PostgresSpec struct {
	DbName   string `json:"dbName,omitempty" protobuf:"bytes,1,opt,name=dbName"`
	Host     string `json:"host,omitempty" protobuf:"bytes,2,opt,name=host"`
	Port     int64  `json:"port,omitempty" protobuf:"varint,3,opt,name=port"`
	User     string `json:"user,omitempty" protobuf:"bytes,4,opt,name=user"`
	Password string `json:"password,omitempty" protobuf:"bytes,5,opt,name=password"`
}

type StorageBackend struct {
	CredentialName string `json:"credentialName,omitempty" protobuf:"bytes,1,opt,name=credentialName"`

	Local    *LocalSpec        `json:"local,omitempty" protobuf:"bytes,2,opt,name=local"`
	S3       *S3Spec           `json:"s3,omitempty" protobuf:"bytes,3,opt,name=s3"`
	GCS      *GCSSpec          `json:"gcs,omitempty" protobuf:"bytes,4,opt,name=gcs"`
	Azure    *AzureStorageSpec `json:"azure,omitempty" protobuf:"bytes,5,opt,name=azure"`
	Swift    *SwiftSpec        `json:"swift,omitempty" protobuf:"bytes,6,opt,name=swift"`
	Postgres *PostgresSpec     `json:"postgres,omitempty" protobuf:"bytes,7,opt,name=postgres"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PharmerConfig struct {
	metav1.TypeMeta `json:",inline,omitempty,omitempty"`
	Context         string         `json:"context,omitempty" protobuf:"bytes,1,opt,name=context"`
	Credentials     []Credential   `json:"credentials,omitempty" protobuf:"bytes,2,rep,name=credentials"`
	Store           StorageBackend `json:"store,omitempty" protobuf:"bytes,3,opt,name=store"`
}

func (pc PharmerConfig) GetStoreType() string {
	if pc.Store.Local != nil {
		return "Local"
	} else if pc.Store.S3 != nil {
		return "S3"
	} else if pc.Store.GCS != nil {
		return "GCS"
	} else if pc.Store.Azure != nil {
		return "Azure"
	} else if pc.Store.Swift != nil {
		return "OpenStack Swift"
	} else if pc.Store.Postgres != nil {
		return "Postgres"
	}
	return "<Unknown>"
}

func (pc PharmerConfig) GetCredential(name string) (*Credential, error) {
	for _, c := range pc.Credentials {
		if c.Name == name {
			return &c, nil
		}
	}
	return nil, errors.Errorf("Missing credential %s", name)
}

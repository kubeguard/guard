package v1alpha1

// Auto-generated. DO NOT EDIT.
import (
	"github.com/golang/glog"
	"github.com/xeipuuv/gojsonschema"
)

var secretEditRequestSchema *gojsonschema.Schema
var persistentVolumeClaimRegisterRequestSchema *gojsonschema.Schema
var diskListRequestSchema *gojsonschema.Schema
var createResourceRequestSchema *gojsonschema.Schema
var updateResourceRequestSchema *gojsonschema.Schema
var diskDescribeRequestSchema *gojsonschema.Schema
var persistentVolumeUnRegisterRequestSchema *gojsonschema.Schema
var copyResourceRequestSchema *gojsonschema.Schema
var deleteResourceRequestSchema *gojsonschema.Schema
var configMapEditRequestSchema *gojsonschema.Schema
var listResourceRequestSchema *gojsonschema.Schema
var persistentVolumeClaimUnRegisterRequestSchema *gojsonschema.Schema
var diskCreateRequestSchema *gojsonschema.Schema
var diskDeleteRequestSchema *gojsonschema.Schema
var persistentVolumeRegisterRequestSchema *gojsonschema.Schema
var describeResourceRequestSchema *gojsonschema.Schema
var reverseIndexResourceRequestSchema *gojsonschema.Schema

func init() {
	var err error
	secretEditRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "add": {
      "additionalProperties": {
        "type": "string"
      },
      "type": "object"
    },
    "cluster": {
      "type": "string"
    },
    "deleted": {
      "items": {
        "type": "string"
      },
      "type": "array"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "namespace": {
      "maxLength": 63,
      "pattern": "^[a-z0-9](?:[a-z0-9\\-]{0,61}[a-z0-9])?$",
      "type": "string"
    },
    "update": {
      "additionalProperties": {
        "type": "string"
      },
      "type": "object"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	persistentVolumeClaimRegisterRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "cluster": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "namespace": {
      "maxLength": 63,
      "pattern": "^[a-z0-9](?:[a-z0-9\\-]{0,61}[a-z0-9])?$",
      "type": "string"
    },
    "size_gb": {
      "type": "integer"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	diskListRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "cluster": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	createResourceRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "v1alpha1Raw": {
      "properties": {
        "data": {
          "type": "string"
        },
        "format": {
          "type": "string"
        }
      },
      "type": "object"
    }
  },
  "properties": {
    "cluster": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "raw": {
      "$ref": "#/definitions/v1alpha1Raw"
    },
    "type": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	updateResourceRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "v1alpha1Raw": {
      "properties": {
        "data": {
          "type": "string"
        },
        "format": {
          "type": "string"
        }
      },
      "type": "object"
    }
  },
  "properties": {
    "cluster": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "raw": {
      "$ref": "#/definitions/v1alpha1Raw"
    },
    "type": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	diskDescribeRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "cluster": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "provider": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	persistentVolumeUnRegisterRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "cluster": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	copyResourceRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "v1alpha1KubeObject": {
      "properties": {
        "cluster": {
          "type": "string"
        },
        "name": {
          "maxLength": 63,
          "type": "string"
        },
        "namespace": {
          "maxLength": 63,
          "pattern": "^[a-z0-9](?:[a-z0-9\\-]{0,61}[a-z0-9])?$",
          "type": "string"
        },
        "type": {
          "type": "string"
        }
      },
      "type": "object"
    }
  },
  "properties": {
    "api_version": {
      "type": "string"
    },
    "destination": {
      "$ref": "#/definitions/v1alpha1KubeObject"
    },
    "source": {
      "$ref": "#/definitions/v1alpha1KubeObject"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	deleteResourceRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "api_version": {
      "type": "string"
    },
    "cluster": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "namespace": {
      "maxLength": 63,
      "type": "string"
    },
    "type": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	configMapEditRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "add": {
      "additionalProperties": {
        "type": "string"
      },
      "type": "object"
    },
    "cluster": {
      "type": "string"
    },
    "deleted": {
      "items": {
        "type": "string"
      },
      "type": "array"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "namespace": {
      "maxLength": 63,
      "pattern": "^[a-z0-9](?:[a-z0-9\\-]{0,61}[a-z0-9])?$",
      "type": "string"
    },
    "update": {
      "additionalProperties": {
        "type": "string"
      },
      "type": "object"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	listResourceRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "api_version": {
      "type": "string"
    },
    "cluster": {
      "type": "string"
    },
    "fieldSelector": {
      "type": "string"
    },
    "include_metrics": {
      "type": "boolean"
    },
    "namespace": {
      "maxLength": 63,
      "pattern": "^[a-z0-9](?:[a-z0-9\\-]{0,61}[a-z0-9])?$",
      "type": "string"
    },
    "selector": {
      "title": "map type is not supported by grpc-gateway as query params.\nhttps://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/query.go#L57\nhttps://github.com/grpc-ecosystem/grpc-gateway/issues/316\nmap<string, string> label_selector = 6;\nexample label_selector=environment=production,tier=frontend",
      "type": "string"
    },
    "type": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	persistentVolumeClaimUnRegisterRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "cluster": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "namespace": {
      "maxLength": 63,
      "pattern": "^[a-z0-9](?:[a-z0-9\\-]{0,61}[a-z0-9])?$",
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	diskCreateRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "cluster": {
      "type": "string"
    },
    "disk_type": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "size_gb": {
      "type": "integer"
    },
    "zone": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	diskDeleteRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "cluster": {
      "type": "string"
    },
    "uid": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	persistentVolumeRegisterRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "cluster": {
      "type": "string"
    },
    "endpoint": {
      "type": "string"
    },
    "identifier": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "plugin": {
      "type": "string"
    },
    "size_gb": {
      "type": "integer"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	describeResourceRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "api_version": {
      "type": "string"
    },
    "cluster": {
      "type": "string"
    },
    "include_metrics": {
      "type": "boolean"
    },
    "name": {
      "maxLength": 63,
      "type": "string"
    },
    "namespace": {
      "maxLength": 63,
      "pattern": "^[a-z0-9](?:[a-z0-9\\-]{0,61}[a-z0-9])?$",
      "type": "string"
    },
    "raw": {
      "type": "string"
    },
    "type": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	reverseIndexResourceRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "api_version": {
      "type": "string"
    },
    "cluster": {
      "type": "string"
    },
    "name": {
      "maxLength": 63,
      "pattern": "^[a-z0-9](?:[a-z0-9\\-]{0,61}[a-z0-9])?$",
      "type": "string"
    },
    "namespace": {
      "maxLength": 63,
      "pattern": "^[a-z0-9](?:[a-z0-9\\-]{0,61}[a-z0-9])?$",
      "type": "string"
    },
    "targetType": {
      "type": "string"
    },
    "type": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
}

func (m *SecretEditRequest) Valid() (*gojsonschema.Result, error) {
	return secretEditRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *SecretEditRequest) IsRequest() {}

func (m *PersistentVolumeClaimRegisterRequest) Valid() (*gojsonschema.Result, error) {
	return persistentVolumeClaimRegisterRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *PersistentVolumeClaimRegisterRequest) IsRequest() {}

func (m *DiskListRequest) Valid() (*gojsonschema.Result, error) {
	return diskListRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *DiskListRequest) IsRequest() {}

func (m *CreateResourceRequest) Valid() (*gojsonschema.Result, error) {
	return createResourceRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *CreateResourceRequest) IsRequest() {}

func (m *UpdateResourceRequest) Valid() (*gojsonschema.Result, error) {
	return updateResourceRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *UpdateResourceRequest) IsRequest() {}

func (m *DiskDescribeRequest) Valid() (*gojsonschema.Result, error) {
	return diskDescribeRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *DiskDescribeRequest) IsRequest() {}

func (m *PersistentVolumeUnRegisterRequest) Valid() (*gojsonschema.Result, error) {
	return persistentVolumeUnRegisterRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *PersistentVolumeUnRegisterRequest) IsRequest() {}

func (m *CopyResourceRequest) Valid() (*gojsonschema.Result, error) {
	return copyResourceRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *CopyResourceRequest) IsRequest() {}

func (m *DeleteResourceRequest) Valid() (*gojsonschema.Result, error) {
	return deleteResourceRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *DeleteResourceRequest) IsRequest() {}

func (m *ConfigMapEditRequest) Valid() (*gojsonschema.Result, error) {
	return configMapEditRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ConfigMapEditRequest) IsRequest() {}

func (m *ListResourceRequest) Valid() (*gojsonschema.Result, error) {
	return listResourceRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ListResourceRequest) IsRequest() {}

func (m *PersistentVolumeClaimUnRegisterRequest) Valid() (*gojsonschema.Result, error) {
	return persistentVolumeClaimUnRegisterRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *PersistentVolumeClaimUnRegisterRequest) IsRequest() {}

func (m *DiskCreateRequest) Valid() (*gojsonschema.Result, error) {
	return diskCreateRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *DiskCreateRequest) IsRequest() {}

func (m *DiskDeleteRequest) Valid() (*gojsonschema.Result, error) {
	return diskDeleteRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *DiskDeleteRequest) IsRequest() {}

func (m *PersistentVolumeRegisterRequest) Valid() (*gojsonschema.Result, error) {
	return persistentVolumeRegisterRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *PersistentVolumeRegisterRequest) IsRequest() {}

func (m *DescribeResourceRequest) Valid() (*gojsonschema.Result, error) {
	return describeResourceRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *DescribeResourceRequest) IsRequest() {}

func (m *ReverseIndexResourceRequest) Valid() (*gojsonschema.Result, error) {
	return reverseIndexResourceRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ReverseIndexResourceRequest) IsRequest() {}


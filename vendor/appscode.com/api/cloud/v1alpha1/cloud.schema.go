package v1alpha1

// Auto-generated. DO NOT EDIT.
import (
	"github.com/golang/glog"
	"github.com/xeipuuv/gojsonschema"
)

var sSHConfigGetRequestSchema *gojsonschema.Schema
var clusterApplyRequestSchema *gojsonschema.Schema
var credentialCreateRequestSchema *gojsonschema.Schema
var nodeGroupDeleteRequestSchema *gojsonschema.Schema
var clusterUpdateRequestSchema *gojsonschema.Schema
var nodeGroupListRequestSchema *gojsonschema.Schema
var clusterDeleteRequestSchema *gojsonschema.Schema
var nodeGroupDescribeRequestSchema *gojsonschema.Schema
var credentialDescribeRequestSchema *gojsonschema.Schema
var credentialDeleteRequestSchema *gojsonschema.Schema
var clusterDescribeRequestSchema *gojsonschema.Schema
var clusterListRequestSchema *gojsonschema.Schema
var clusterClientConfigRequestSchema *gojsonschema.Schema
var nodeGroupUpdateRequestSchema *gojsonschema.Schema
var nodeGroupCreateRequestSchema *gojsonschema.Schema
var credentialUpdateRequestSchema *gojsonschema.Schema
var clusterCreateRequestSchema *gojsonschema.Schema
var clusterMetadataRequestSchema *gojsonschema.Schema

func init() {
	var err error
	sSHConfigGetRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "clusterName": {
      "type": "string"
    },
    "nodeName": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	clusterApplyRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "name": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	credentialCreateRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "apismetav1ObjectMeta": {
      "description": "ObjectMeta is metadata that all persisted resources must have, which includes all objects\nusers must create.",
      "properties": {
        "annotations": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Annotations is an unstructured key value map stored with a resource that may be\nset by external tools to store and retrieve arbitrary metadata. They are not\nqueryable and should be preserved when modifying objects.\nMore info: http://kubernetes.io/docs/user-guide/annotations\n+optional",
          "type": "object"
        },
        "clusterName": {
          "title": "The name of the cluster which the object belongs to.\nThis is used to distinguish resources with same name and namespace in different clusters.\nThis field is not set anywhere right now and apiserver is going to ignore it if set in create or update request.\n+optional",
          "type": "string"
        },
        "creationTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "CreationTimestamp is a timestamp representing the server time when this object was\ncreated. It is not guaranteed to be set in happens-before order across separate operations.\nClients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system.\nRead-only.\nNull for lists.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "deletionGracePeriodSeconds": {
          "title": "Number of seconds allowed for this object to gracefully terminate before\nit will be removed from the system. Only set when deletionTimestamp is also set.\nMay only be shortened.\nRead-only.\n+optional",
          "type": "integer"
        },
        "deletionTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted. This\nfield is set by the server when a graceful deletion is requested by the user, and is not\ndirectly settable by a client. The resource is expected to be deleted (no longer visible\nfrom resource lists, and not reachable by name) after the time in this field, once the\nfinalizers list is empty. As long as the finalizers list contains items, deletion is blocked.\nOnce the deletionTimestamp is set, this value may not be unset or be set further into the\nfuture, although it may be shortened or the resource may be deleted prior to this time.\nFor example, a user may request that a pod is deleted in 30 seconds. The Kubelet will react\nby sending a graceful termination signal to the containers in the pod. After that 30 seconds,\nthe Kubelet will send a hard termination signal (SIGKILL) to the container and after cleanup,\nremove the pod from the API. In the presence of network partitions, this object may still\nexist after this timestamp, until an administrator or automated process can determine the\nresource is fully terminated.\nIf not set, graceful deletion of the object has not been requested.\n\nPopulated by the system when a graceful deletion is requested.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "finalizers": {
          "items": {
            "type": "string"
          },
          "title": "Must be empty before the object is deleted from the registry. Each entry\nis an identifier for the responsible component that will remove the entry\nfrom the list. If the deletionTimestamp of the object is non-nil, entries\nin this list can only be removed.\n+optional\n+patchStrategy=merge",
          "type": "array"
        },
        "generateName": {
          "description": "GenerateName is an optional prefix, used by the server, to generate a unique\nname ONLY IF the Name field has not been provided.\nIf this field is used, the name returned to the client will be different\nthan the name passed. This value will also be combined with a unique suffix.\nThe provided value has the same validation rules as the Name field,\nand may be truncated by the length of the suffix required to make the value\nunique on the server.\n\nIf this field is specified and the generated name exists, the server will\nNOT return a 409 - instead, it will either return 201 Created or 500 with Reason\nServerTimeout indicating a unique name could not be found in the time allotted, and the client\nshould retry (optionally after the time indicated in the Retry-After header).\n\nApplied only if Name is not specified.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency\n+optional",
          "type": "string"
        },
        "generation": {
          "title": "A sequence number representing a specific generation of the desired state.\nPopulated by the system. Read-only.\n+optional",
          "type": "integer"
        },
        "initializers": {
          "$ref": "#/definitions/v1Initializers",
          "description": "An initializer is a controller which enforces some system invariant at object creation time.\nThis field is a list of initializers that have not yet acted on this object. If nil or empty,\nthis object has been completely initialized. Otherwise, the object is considered uninitialized\nand is hidden (in list/watch and get calls) from clients that haven't explicitly asked to\nobserve uninitialized objects.\n\nWhen an object is created, the system will populate this list with the current set of initializers.\nOnly privileged users may set or modify this list. Once it is empty, it may not be modified further\nby any user."
        },
        "labels": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Map of string keys and values that can be used to organize and categorize\n(scope and select) objects. May match selectors of replication controllers\nand services.\nMore info: http://kubernetes.io/docs/user-guide/labels\n+optional",
          "type": "object"
        },
        "name": {
          "title": "Name must be unique within a namespace. Is required when creating resources, although\nsome resources may allow a client to request the generation of an appropriate name\nautomatically. Name is primarily intended for creation idempotence and configuration\ndefinition.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names\n+optional",
          "type": "string"
        },
        "namespace": {
          "description": "Namespace defines the space within each name must be unique. An empty namespace is\nequivalent to the \"default\" namespace, but \"default\" is the canonical representation.\nNot all objects are required to be scoped to a namespace - the value of this field for\nthose objects will be empty.\n\nMust be a DNS_LABEL.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/namespaces\n+optional",
          "type": "string"
        },
        "ownerReferences": {
          "items": {
            "$ref": "#/definitions/v1OwnerReference"
          },
          "title": "List of objects depended by this object. If ALL objects in the list have\nbeen deleted, this object will be garbage collected. If this object is managed by a controller,\nthen an entry in this list will point to this controller, with the controller field set to true.\nThere cannot be more than one managing controller.\n+optional\n+patchMergeKey=uid\n+patchStrategy=merge",
          "type": "array"
        },
        "resourceVersion": {
          "description": "An opaque value that represents the internal version of this object that can\nbe used by clients to determine when objects have changed. May be used for optimistic\nconcurrency, change detection, and the watch operation on a resource or set of resources.\nClients must treat these values as opaque and passed unmodified back to the server.\nThey may only be valid for a particular resource or set of resources.\n\nPopulated by the system.\nRead-only.\nValue must be treated as opaque by clients and .\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "SelfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        },
        "uid": {
          "description": "UID is the unique in time and space value for this object. It is typically generated by\nthe server on successful creation of a resource and is not allowed to change on PUT\noperations.\n\nPopulated by the system.\nRead-only.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializer": {
      "description": "Initializer is information about an initializer that has not yet completed.",
      "properties": {
        "name": {
          "description": "name of the process that is responsible for initializing this object.",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializers": {
      "description": "Initializers tracks the progress of initialization.",
      "properties": {
        "pending": {
          "items": {
            "$ref": "#/definitions/v1Initializer"
          },
          "title": "Pending is a list of initializers that must execute in order before this object is visible.\nWhen the last pending initializer is removed, and no failing result is set, the initializers\nstruct will be set to nil and the object is considered as initialized and visible to all\nclients.\n+patchMergeKey=name\n+patchStrategy=merge",
          "type": "array"
        },
        "result": {
          "$ref": "#/definitions/v1Status",
          "description": "If result is set with the Failure field, the object will be persisted to storage and then deleted,\nensuring that other clients can observe the deletion."
        }
      },
      "type": "object"
    },
    "v1ListMeta": {
      "description": "ListMeta describes metadata that synthetic resources must have, including lists and\nvarious status objects. A resource may have only one of {ObjectMeta, ListMeta}.",
      "properties": {
        "continue": {
          "description": "continue may be set if the user set a limit on the number of items returned, and indicates that\nthe server has more data available. The value is opaque and may be used to issue another request\nto the endpoint that served this list to retrieve the next set of available objects. Continuing a\nlist may not be possible if the server configuration has changed or more than a few minutes have\npassed. The resourceVersion field returned when using this continue value will be identical to\nthe value in the first response.",
          "type": "string"
        },
        "resourceVersion": {
          "title": "String that identifies the server's internal version of this object that\ncan be used by clients to determine when objects have changed.\nValue must be treated as opaque by clients and passed unmodified back to the server.\nPopulated by the system.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "selfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1OwnerReference": {
      "description": "OwnerReference contains enough information to let you identify an owning\nobject. Currently, an owning object must be in the same namespace, so there\nis no namespace field.",
      "properties": {
        "apiVersion": {
          "description": "API version of the referent.",
          "type": "string"
        },
        "blockOwnerDeletion": {
          "title": "If true, AND if the owner has the \"foregroundDeletion\" finalizer, then\nthe owner cannot be deleted from the key-value store until this\nreference is removed.\nDefaults to false.\nTo set this field, a user needs \"delete\" permission of the owner,\notherwise 422 (Unprocessable Entity) will be returned.\n+optional",
          "type": "boolean"
        },
        "controller": {
          "title": "If true, this reference points to the managing controller.\n+optional",
          "type": "boolean"
        },
        "kind": {
          "title": "Kind of the referent.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
          "type": "string"
        },
        "name": {
          "title": "Name of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names",
          "type": "string"
        },
        "uid": {
          "title": "UID of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Status": {
      "description": "Status is a return value for calls that don't return other objects.",
      "properties": {
        "code": {
          "title": "Suggested HTTP return code for this status, 0 if not set.\n+optional",
          "type": "integer"
        },
        "details": {
          "$ref": "#/definitions/v1StatusDetails",
          "title": "Extended data associated with the reason.  Each reason may define its\nown extended details. This field is optional and the data returned\nis not guaranteed to conform to any schema except that defined by\nthe reason type.\n+optional"
        },
        "message": {
          "title": "A human-readable description of the status of this operation.\n+optional",
          "type": "string"
        },
        "metadata": {
          "$ref": "#/definitions/v1ListMeta",
          "title": "Standard list metadata.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional"
        },
        "reason": {
          "title": "A machine-readable description of why this operation is in the\n\"Failure\" status. If this value is empty there\nis no information available. A Reason clarifies an HTTP status\ncode but does not override it.\n+optional",
          "type": "string"
        },
        "status": {
          "title": "Status of the operation.\nOne of: \"Success\" or \"Failure\".\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusCause": {
      "description": "StatusCause provides more information about an api.Status failure, including\ncases when multiple errors are encountered.",
      "properties": {
        "field": {
          "description": "The field of the resource that has caused this error, as named by its JSON\nserialization. May include dot and postfix notation for nested attributes.\nArrays are zero-indexed.  Fields may appear more than once in an array of\ncauses due to fields having multiple errors.\nOptional.\n\nExamples:\n  \"name\" - the field \"name\" on the current resource\n  \"items[0].name\" - the field \"name\" on the first array entry in \"items\"\n+optional",
          "type": "string"
        },
        "message": {
          "title": "A human-readable description of the cause of the error.  This field may be\npresented as-is to a reader.\n+optional",
          "type": "string"
        },
        "reason": {
          "title": "A machine-readable description of the cause of the error. If this value is\nempty there is no information available.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusDetails": {
      "description": "StatusDetails is a set of additional properties that MAY be set by the\nserver to provide additional information about a response. The Reason\nfield of a Status object defines what attributes will be set. Clients\nmust ignore fields that do not match the defined type of each attribute,\nand should assume that any attribute may be empty, invalid, or under\ndefined.",
      "properties": {
        "causes": {
          "items": {
            "$ref": "#/definitions/v1StatusCause"
          },
          "title": "The Causes array includes more details associated with the StatusReason\nfailure. Not all StatusReasons may provide detailed causes.\n+optional",
          "type": "array"
        },
        "group": {
          "title": "The group attribute of the resource associated with the status StatusReason.\n+optional",
          "type": "string"
        },
        "kind": {
          "title": "The kind attribute of the resource associated with the status StatusReason.\nOn some operations may differ from the requested resource Kind.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional",
          "type": "string"
        },
        "name": {
          "title": "The name attribute of the resource associated with the status StatusReason\n(when there is a single name which can be described).\n+optional",
          "type": "string"
        },
        "retryAfterSeconds": {
          "title": "If specified, the time in seconds before the operation should be retried. Some errors may indicate\nthe client must take an alternate action - for those errors this field may indicate how long to wait\nbefore taking the alternate action.\n+optional",
          "type": "integer"
        },
        "uid": {
          "title": "UID of the resource.\n(when there is a single resource which can be described).\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Time": {
      "description": "Time is a wrapper around time.Time which supports correct\nmarshaling to YAML and JSON.  Wrappers are provided for many\nof the factory methods that the time package offers.\n\n+protobuf.options.marshal=false\n+protobuf.as=Timestamp\n+protobuf.options.(gogoproto.goproto_stringer)=false",
      "properties": {
        "nanos": {
          "description": "Non-negative fractions of a second at nanosecond resolution. Negative\nsecond values with fractions must still have non-negative nanos values\nthat count forward in time. Must be from 0 to 999,999,999\ninclusive. This field may be limited in precision depending on context.",
          "type": "integer"
        },
        "seconds": {
          "description": "Represents seconds of UTC time since Unix epoch\n1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to\n9999-12-31T23:59:59Z inclusive.",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1Credential": {
      "properties": {
        "metadata": {
          "$ref": "#/definitions/apismetav1ObjectMeta"
        },
        "spec": {
          "$ref": "#/definitions/v1alpha1CredentialSpec"
        }
      },
      "type": "object"
    },
    "v1alpha1CredentialSpec": {
      "properties": {
        "data": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "provider": {
          "type": "string"
        }
      },
      "type": "object"
    }
  },
  "properties": {
    "credential": {
      "$ref": "#/definitions/v1alpha1Credential"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	nodeGroupDeleteRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "clusterName": {
      "type": "string"
    },
    "name": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	clusterUpdateRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "apismetav1ObjectMeta": {
      "description": "ObjectMeta is metadata that all persisted resources must have, which includes all objects\nusers must create.",
      "properties": {
        "annotations": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Annotations is an unstructured key value map stored with a resource that may be\nset by external tools to store and retrieve arbitrary metadata. They are not\nqueryable and should be preserved when modifying objects.\nMore info: http://kubernetes.io/docs/user-guide/annotations\n+optional",
          "type": "object"
        },
        "clusterName": {
          "title": "The name of the cluster which the object belongs to.\nThis is used to distinguish resources with same name and namespace in different clusters.\nThis field is not set anywhere right now and apiserver is going to ignore it if set in create or update request.\n+optional",
          "type": "string"
        },
        "creationTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "CreationTimestamp is a timestamp representing the server time when this object was\ncreated. It is not guaranteed to be set in happens-before order across separate operations.\nClients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system.\nRead-only.\nNull for lists.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "deletionGracePeriodSeconds": {
          "title": "Number of seconds allowed for this object to gracefully terminate before\nit will be removed from the system. Only set when deletionTimestamp is also set.\nMay only be shortened.\nRead-only.\n+optional",
          "type": "integer"
        },
        "deletionTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted. This\nfield is set by the server when a graceful deletion is requested by the user, and is not\ndirectly settable by a client. The resource is expected to be deleted (no longer visible\nfrom resource lists, and not reachable by name) after the time in this field, once the\nfinalizers list is empty. As long as the finalizers list contains items, deletion is blocked.\nOnce the deletionTimestamp is set, this value may not be unset or be set further into the\nfuture, although it may be shortened or the resource may be deleted prior to this time.\nFor example, a user may request that a pod is deleted in 30 seconds. The Kubelet will react\nby sending a graceful termination signal to the containers in the pod. After that 30 seconds,\nthe Kubelet will send a hard termination signal (SIGKILL) to the container and after cleanup,\nremove the pod from the API. In the presence of network partitions, this object may still\nexist after this timestamp, until an administrator or automated process can determine the\nresource is fully terminated.\nIf not set, graceful deletion of the object has not been requested.\n\nPopulated by the system when a graceful deletion is requested.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "finalizers": {
          "items": {
            "type": "string"
          },
          "title": "Must be empty before the object is deleted from the registry. Each entry\nis an identifier for the responsible component that will remove the entry\nfrom the list. If the deletionTimestamp of the object is non-nil, entries\nin this list can only be removed.\n+optional\n+patchStrategy=merge",
          "type": "array"
        },
        "generateName": {
          "description": "GenerateName is an optional prefix, used by the server, to generate a unique\nname ONLY IF the Name field has not been provided.\nIf this field is used, the name returned to the client will be different\nthan the name passed. This value will also be combined with a unique suffix.\nThe provided value has the same validation rules as the Name field,\nand may be truncated by the length of the suffix required to make the value\nunique on the server.\n\nIf this field is specified and the generated name exists, the server will\nNOT return a 409 - instead, it will either return 201 Created or 500 with Reason\nServerTimeout indicating a unique name could not be found in the time allotted, and the client\nshould retry (optionally after the time indicated in the Retry-After header).\n\nApplied only if Name is not specified.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency\n+optional",
          "type": "string"
        },
        "generation": {
          "title": "A sequence number representing a specific generation of the desired state.\nPopulated by the system. Read-only.\n+optional",
          "type": "integer"
        },
        "initializers": {
          "$ref": "#/definitions/v1Initializers",
          "description": "An initializer is a controller which enforces some system invariant at object creation time.\nThis field is a list of initializers that have not yet acted on this object. If nil or empty,\nthis object has been completely initialized. Otherwise, the object is considered uninitialized\nand is hidden (in list/watch and get calls) from clients that haven't explicitly asked to\nobserve uninitialized objects.\n\nWhen an object is created, the system will populate this list with the current set of initializers.\nOnly privileged users may set or modify this list. Once it is empty, it may not be modified further\nby any user."
        },
        "labels": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Map of string keys and values that can be used to organize and categorize\n(scope and select) objects. May match selectors of replication controllers\nand services.\nMore info: http://kubernetes.io/docs/user-guide/labels\n+optional",
          "type": "object"
        },
        "name": {
          "title": "Name must be unique within a namespace. Is required when creating resources, although\nsome resources may allow a client to request the generation of an appropriate name\nautomatically. Name is primarily intended for creation idempotence and configuration\ndefinition.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names\n+optional",
          "type": "string"
        },
        "namespace": {
          "description": "Namespace defines the space within each name must be unique. An empty namespace is\nequivalent to the \"default\" namespace, but \"default\" is the canonical representation.\nNot all objects are required to be scoped to a namespace - the value of this field for\nthose objects will be empty.\n\nMust be a DNS_LABEL.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/namespaces\n+optional",
          "type": "string"
        },
        "ownerReferences": {
          "items": {
            "$ref": "#/definitions/v1OwnerReference"
          },
          "title": "List of objects depended by this object. If ALL objects in the list have\nbeen deleted, this object will be garbage collected. If this object is managed by a controller,\nthen an entry in this list will point to this controller, with the controller field set to true.\nThere cannot be more than one managing controller.\n+optional\n+patchMergeKey=uid\n+patchStrategy=merge",
          "type": "array"
        },
        "resourceVersion": {
          "description": "An opaque value that represents the internal version of this object that can\nbe used by clients to determine when objects have changed. May be used for optimistic\nconcurrency, change detection, and the watch operation on a resource or set of resources.\nClients must treat these values as opaque and passed unmodified back to the server.\nThey may only be valid for a particular resource or set of resources.\n\nPopulated by the system.\nRead-only.\nValue must be treated as opaque by clients and .\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "SelfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        },
        "uid": {
          "description": "UID is the unique in time and space value for this object. It is typically generated by\nthe server on successful creation of a resource and is not allowed to change on PUT\noperations.\n\nPopulated by the system.\nRead-only.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializer": {
      "description": "Initializer is information about an initializer that has not yet completed.",
      "properties": {
        "name": {
          "description": "name of the process that is responsible for initializing this object.",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializers": {
      "description": "Initializers tracks the progress of initialization.",
      "properties": {
        "pending": {
          "items": {
            "$ref": "#/definitions/v1Initializer"
          },
          "title": "Pending is a list of initializers that must execute in order before this object is visible.\nWhen the last pending initializer is removed, and no failing result is set, the initializers\nstruct will be set to nil and the object is considered as initialized and visible to all\nclients.\n+patchMergeKey=name\n+patchStrategy=merge",
          "type": "array"
        },
        "result": {
          "$ref": "#/definitions/v1Status",
          "description": "If result is set with the Failure field, the object will be persisted to storage and then deleted,\nensuring that other clients can observe the deletion."
        }
      },
      "type": "object"
    },
    "v1ListMeta": {
      "description": "ListMeta describes metadata that synthetic resources must have, including lists and\nvarious status objects. A resource may have only one of {ObjectMeta, ListMeta}.",
      "properties": {
        "continue": {
          "description": "continue may be set if the user set a limit on the number of items returned, and indicates that\nthe server has more data available. The value is opaque and may be used to issue another request\nto the endpoint that served this list to retrieve the next set of available objects. Continuing a\nlist may not be possible if the server configuration has changed or more than a few minutes have\npassed. The resourceVersion field returned when using this continue value will be identical to\nthe value in the first response.",
          "type": "string"
        },
        "resourceVersion": {
          "title": "String that identifies the server's internal version of this object that\ncan be used by clients to determine when objects have changed.\nValue must be treated as opaque by clients and passed unmodified back to the server.\nPopulated by the system.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "selfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1NodeAddress": {
      "description": "NodeAddress contains information for the node's address.",
      "properties": {
        "address": {
          "description": "The node address.",
          "type": "string"
        },
        "type": {
          "description": "Node address type, one of Hostname, ExternalIP or InternalIP.",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1OwnerReference": {
      "description": "OwnerReference contains enough information to let you identify an owning\nobject. Currently, an owning object must be in the same namespace, so there\nis no namespace field.",
      "properties": {
        "apiVersion": {
          "description": "API version of the referent.",
          "type": "string"
        },
        "blockOwnerDeletion": {
          "title": "If true, AND if the owner has the \"foregroundDeletion\" finalizer, then\nthe owner cannot be deleted from the key-value store until this\nreference is removed.\nDefaults to false.\nTo set this field, a user needs \"delete\" permission of the owner,\notherwise 422 (Unprocessable Entity) will be returned.\n+optional",
          "type": "boolean"
        },
        "controller": {
          "title": "If true, this reference points to the managing controller.\n+optional",
          "type": "boolean"
        },
        "kind": {
          "title": "Kind of the referent.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
          "type": "string"
        },
        "name": {
          "title": "Name of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names",
          "type": "string"
        },
        "uid": {
          "title": "UID of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Status": {
      "description": "Status is a return value for calls that don't return other objects.",
      "properties": {
        "code": {
          "title": "Suggested HTTP return code for this status, 0 if not set.\n+optional",
          "type": "integer"
        },
        "details": {
          "$ref": "#/definitions/v1StatusDetails",
          "title": "Extended data associated with the reason.  Each reason may define its\nown extended details. This field is optional and the data returned\nis not guaranteed to conform to any schema except that defined by\nthe reason type.\n+optional"
        },
        "message": {
          "title": "A human-readable description of the status of this operation.\n+optional",
          "type": "string"
        },
        "metadata": {
          "$ref": "#/definitions/v1ListMeta",
          "title": "Standard list metadata.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional"
        },
        "reason": {
          "title": "A machine-readable description of why this operation is in the\n\"Failure\" status. If this value is empty there\nis no information available. A Reason clarifies an HTTP status\ncode but does not override it.\n+optional",
          "type": "string"
        },
        "status": {
          "title": "Status of the operation.\nOne of: \"Success\" or \"Failure\".\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusCause": {
      "description": "StatusCause provides more information about an api.Status failure, including\ncases when multiple errors are encountered.",
      "properties": {
        "field": {
          "description": "The field of the resource that has caused this error, as named by its JSON\nserialization. May include dot and postfix notation for nested attributes.\nArrays are zero-indexed.  Fields may appear more than once in an array of\ncauses due to fields having multiple errors.\nOptional.\n\nExamples:\n  \"name\" - the field \"name\" on the current resource\n  \"items[0].name\" - the field \"name\" on the first array entry in \"items\"\n+optional",
          "type": "string"
        },
        "message": {
          "title": "A human-readable description of the cause of the error.  This field may be\npresented as-is to a reader.\n+optional",
          "type": "string"
        },
        "reason": {
          "title": "A machine-readable description of the cause of the error. If this value is\nempty there is no information available.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusDetails": {
      "description": "StatusDetails is a set of additional properties that MAY be set by the\nserver to provide additional information about a response. The Reason\nfield of a Status object defines what attributes will be set. Clients\nmust ignore fields that do not match the defined type of each attribute,\nand should assume that any attribute may be empty, invalid, or under\ndefined.",
      "properties": {
        "causes": {
          "items": {
            "$ref": "#/definitions/v1StatusCause"
          },
          "title": "The Causes array includes more details associated with the StatusReason\nfailure. Not all StatusReasons may provide detailed causes.\n+optional",
          "type": "array"
        },
        "group": {
          "title": "The group attribute of the resource associated with the status StatusReason.\n+optional",
          "type": "string"
        },
        "kind": {
          "title": "The kind attribute of the resource associated with the status StatusReason.\nOn some operations may differ from the requested resource Kind.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional",
          "type": "string"
        },
        "name": {
          "title": "The name attribute of the resource associated with the status StatusReason\n(when there is a single name which can be described).\n+optional",
          "type": "string"
        },
        "retryAfterSeconds": {
          "title": "If specified, the time in seconds before the operation should be retried. Some errors may indicate\nthe client must take an alternate action - for those errors this field may indicate how long to wait\nbefore taking the alternate action.\n+optional",
          "type": "integer"
        },
        "uid": {
          "title": "UID of the resource.\n(when there is a single resource which can be described).\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Time": {
      "description": "Time is a wrapper around time.Time which supports correct\nmarshaling to YAML and JSON.  Wrappers are provided for many\nof the factory methods that the time package offers.\n\n+protobuf.options.marshal=false\n+protobuf.as=Timestamp\n+protobuf.options.(gogoproto.goproto_stringer)=false",
      "properties": {
        "nanos": {
          "description": "Non-negative fractions of a second at nanosecond resolution. Negative\nsecond values with fractions must still have non-negative nanos values\nthat count forward in time. Must be from 0 to 999,999,999\ninclusive. This field may be limited in precision depending on context.",
          "type": "integer"
        },
        "seconds": {
          "description": "Represents seconds of UTC time since Unix epoch\n1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to\n9999-12-31T23:59:59Z inclusive.",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1API": {
      "properties": {
        "advertiseAddress": {
          "description": "AdvertiseAddress sets the address for the API server to advertise.",
          "type": "string"
        },
        "bindPort": {
          "title": "BindPort sets the secure port for the API Server to bind to",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1AWSSpec": {
      "properties": {
        "iamProfileMaster": {
          "title": "aws:TAG KubernetesCluster => clusterid",
          "type": "string"
        },
        "iamProfileNode": {
          "type": "string"
        },
        "masterIPSuffix": {
          "type": "string"
        },
        "masterSGName": {
          "type": "string"
        },
        "nodeSGName": {
          "type": "string"
        },
        "subnetCidr": {
          "type": "string"
        },
        "vpcCIDR": {
          "type": "string"
        },
        "vpcCIDRBase": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1AWSStatus": {
      "properties": {
        "dhcpOptionsID": {
          "type": "string"
        },
        "igwID": {
          "type": "string"
        },
        "masterSGID": {
          "type": "string"
        },
        "nodeSGID": {
          "type": "string"
        },
        "routeTableID": {
          "type": "string"
        },
        "subnetID": {
          "type": "string"
        },
        "volumeID": {
          "type": "string"
        },
        "vpcID": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1AzureSpec": {
      "properties": {
        "azureStorageAccountName": {
          "type": "string"
        },
        "instanceImageVersion": {
          "type": "string"
        },
        "resourceGroup": {
          "type": "string"
        },
        "rootPassword": {
          "type": "string"
        },
        "routeTableName": {
          "type": "string"
        },
        "securityGroupName": {
          "type": "string"
        },
        "subnetCidr": {
          "type": "string"
        },
        "subnetName": {
          "type": "string"
        },
        "vnetName": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1CloudSpec": {
      "properties": {
        "aws": {
          "$ref": "#/definitions/v1alpha1AWSSpec"
        },
        "azure": {
          "$ref": "#/definitions/v1alpha1AzureSpec"
        },
        "ccmCredentialName": {
          "type": "string"
        },
        "cloudProvider": {
          "type": "string"
        },
        "gce": {
          "$ref": "#/definitions/v1alpha1GoogleSpec"
        },
        "gke": {
          "$ref": "#/definitions/v1alpha1GKESpec"
        },
        "instanceImage": {
          "title": "master needs it for ossec",
          "type": "string"
        },
        "instanceImageProject": {
          "type": "string"
        },
        "linode": {
          "$ref": "#/definitions/v1alpha1LinodeSpec"
        },
        "os": {
          "type": "string"
        },
        "project": {
          "type": "string"
        },
        "region": {
          "type": "string"
        },
        "sshKeyName": {
          "type": "string"
        },
        "zone": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1CloudStatus": {
      "properties": {
        "aws": {
          "$ref": "#/definitions/v1alpha1AWSStatus"
        },
        "sshKeyExternalID": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1Cluster": {
      "properties": {
        "metadata": {
          "$ref": "#/definitions/apismetav1ObjectMeta"
        },
        "spec": {
          "$ref": "#/definitions/v1alpha1ClusterSpec"
        },
        "status": {
          "$ref": "#/definitions/v1alpha1ClusterStatus"
        }
      },
      "type": "object"
    },
    "v1alpha1ClusterSpec": {
      "properties": {
        "api": {
          "$ref": "#/definitions/v1alpha1API"
        },
        "apiServerCertSANs": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "apiServerExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "authorizationModes": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "caCertName": {
          "type": "string"
        },
        "cloud": {
          "$ref": "#/definitions/v1alpha1CloudSpec"
        },
        "controllerManagerExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "credentialName": {
          "type": "string"
        },
        "frontProxyCACertName": {
          "type": "string"
        },
        "kubeletExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "kubernetesVersion": {
          "type": "string"
        },
        "locked": {
          "type": "boolean"
        },
        "networking": {
          "$ref": "#/definitions/v1alpha1Networking"
        },
        "schedulerExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        }
      },
      "type": "object"
    },
    "v1alpha1ClusterStatus": {
      "properties": {
        "apiServer": {
          "items": {
            "$ref": "#/definitions/v1NodeAddress"
          },
          "type": "array"
        },
        "cloud": {
          "$ref": "#/definitions/v1alpha1CloudStatus"
        },
        "phase": {
          "type": "string"
        },
        "reason": {
          "type": "string"
        },
        "reservedIP": {
          "items": {
            "$ref": "#/definitions/v1alpha1ReservedIP"
          },
          "type": "array"
        }
      },
      "type": "object"
    },
    "v1alpha1GKESpec": {
      "properties": {
        "networkName": {
          "type": "string"
        },
        "password": {
          "type": "string"
        },
        "userName": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1GoogleSpec": {
      "properties": {
        "networkName": {
          "type": "string"
        },
        "nodeScopes": {
          "items": {
            "type": "string"
          },
          "title": "gce\nNODE_SCOPES=\"${NODE_SCOPES:-compute-rw,monitoring,logging-write,storage-ro}\"",
          "type": "array"
        },
        "nodeTags": {
          "items": {
            "type": "string"
          },
          "type": "array"
        }
      },
      "type": "object"
    },
    "v1alpha1LinodeSpec": {
      "properties": {
        "kernelId": {
          "type": "integer"
        },
        "rootPassword": {
          "title": "Linode",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1Networking": {
      "properties": {
        "dnsDomain": {
          "type": "string"
        },
        "dnsServerIP": {
          "title": "NEW\nReplacing API_SERVERS https://github.com/kubernetes/kubernetes/blob/62898319dff291843e53b7839c6cde14ee5d2aa4/cluster/aws/util.sh#L1004",
          "type": "string"
        },
        "masterSubnet": {
          "type": "string"
        },
        "networkProvider": {
          "type": "string"
        },
        "nonMasqueradeCIDR": {
          "type": "string"
        },
        "podSubnet": {
          "title": "kubenet, flannel, calico, opencontrail",
          "type": "string"
        },
        "serviceSubnet": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1ReservedIP": {
      "properties": {
        "id": {
          "type": "string"
        },
        "ip": {
          "type": "string"
        },
        "name": {
          "type": "string"
        }
      },
      "type": "object"
    }
  },
  "properties": {
    "cluster": {
      "$ref": "#/definitions/v1alpha1Cluster"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	nodeGroupListRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "clusterName": {
      "type": "string"
    },
    "status": {
      "items": {
        "type": "string"
      },
      "type": "array"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	clusterDeleteRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "deleteDynamicVolumes": {
      "type": "boolean"
    },
    "force": {
      "type": "boolean"
    },
    "keepLodabalancers": {
      "type": "boolean"
    },
    "name": {
      "type": "string"
    },
    "releaseReservedIP": {
      "type": "boolean"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	nodeGroupDescribeRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "clusterName": {
      "type": "string"
    },
    "name": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	credentialDescribeRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "name": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	credentialDeleteRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "delete_dynamic_volumes": {
      "type": "boolean"
    },
    "force": {
      "type": "boolean"
    },
    "keep_lodabalancers": {
      "type": "boolean"
    },
    "name": {
      "type": "string"
    },
    "release_reserved_ip": {
      "type": "boolean"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	clusterDescribeRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "name": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	clusterListRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "status": {
      "items": {
        "type": "string"
      },
      "type": "array"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	clusterClientConfigRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "name": {
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
	nodeGroupUpdateRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "apismetav1ObjectMeta": {
      "description": "ObjectMeta is metadata that all persisted resources must have, which includes all objects\nusers must create.",
      "properties": {
        "annotations": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Annotations is an unstructured key value map stored with a resource that may be\nset by external tools to store and retrieve arbitrary metadata. They are not\nqueryable and should be preserved when modifying objects.\nMore info: http://kubernetes.io/docs/user-guide/annotations\n+optional",
          "type": "object"
        },
        "clusterName": {
          "title": "The name of the cluster which the object belongs to.\nThis is used to distinguish resources with same name and namespace in different clusters.\nThis field is not set anywhere right now and apiserver is going to ignore it if set in create or update request.\n+optional",
          "type": "string"
        },
        "creationTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "CreationTimestamp is a timestamp representing the server time when this object was\ncreated. It is not guaranteed to be set in happens-before order across separate operations.\nClients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system.\nRead-only.\nNull for lists.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "deletionGracePeriodSeconds": {
          "title": "Number of seconds allowed for this object to gracefully terminate before\nit will be removed from the system. Only set when deletionTimestamp is also set.\nMay only be shortened.\nRead-only.\n+optional",
          "type": "integer"
        },
        "deletionTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted. This\nfield is set by the server when a graceful deletion is requested by the user, and is not\ndirectly settable by a client. The resource is expected to be deleted (no longer visible\nfrom resource lists, and not reachable by name) after the time in this field, once the\nfinalizers list is empty. As long as the finalizers list contains items, deletion is blocked.\nOnce the deletionTimestamp is set, this value may not be unset or be set further into the\nfuture, although it may be shortened or the resource may be deleted prior to this time.\nFor example, a user may request that a pod is deleted in 30 seconds. The Kubelet will react\nby sending a graceful termination signal to the containers in the pod. After that 30 seconds,\nthe Kubelet will send a hard termination signal (SIGKILL) to the container and after cleanup,\nremove the pod from the API. In the presence of network partitions, this object may still\nexist after this timestamp, until an administrator or automated process can determine the\nresource is fully terminated.\nIf not set, graceful deletion of the object has not been requested.\n\nPopulated by the system when a graceful deletion is requested.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "finalizers": {
          "items": {
            "type": "string"
          },
          "title": "Must be empty before the object is deleted from the registry. Each entry\nis an identifier for the responsible component that will remove the entry\nfrom the list. If the deletionTimestamp of the object is non-nil, entries\nin this list can only be removed.\n+optional\n+patchStrategy=merge",
          "type": "array"
        },
        "generateName": {
          "description": "GenerateName is an optional prefix, used by the server, to generate a unique\nname ONLY IF the Name field has not been provided.\nIf this field is used, the name returned to the client will be different\nthan the name passed. This value will also be combined with a unique suffix.\nThe provided value has the same validation rules as the Name field,\nand may be truncated by the length of the suffix required to make the value\nunique on the server.\n\nIf this field is specified and the generated name exists, the server will\nNOT return a 409 - instead, it will either return 201 Created or 500 with Reason\nServerTimeout indicating a unique name could not be found in the time allotted, and the client\nshould retry (optionally after the time indicated in the Retry-After header).\n\nApplied only if Name is not specified.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency\n+optional",
          "type": "string"
        },
        "generation": {
          "title": "A sequence number representing a specific generation of the desired state.\nPopulated by the system. Read-only.\n+optional",
          "type": "integer"
        },
        "initializers": {
          "$ref": "#/definitions/v1Initializers",
          "description": "An initializer is a controller which enforces some system invariant at object creation time.\nThis field is a list of initializers that have not yet acted on this object. If nil or empty,\nthis object has been completely initialized. Otherwise, the object is considered uninitialized\nand is hidden (in list/watch and get calls) from clients that haven't explicitly asked to\nobserve uninitialized objects.\n\nWhen an object is created, the system will populate this list with the current set of initializers.\nOnly privileged users may set or modify this list. Once it is empty, it may not be modified further\nby any user."
        },
        "labels": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Map of string keys and values that can be used to organize and categorize\n(scope and select) objects. May match selectors of replication controllers\nand services.\nMore info: http://kubernetes.io/docs/user-guide/labels\n+optional",
          "type": "object"
        },
        "name": {
          "title": "Name must be unique within a namespace. Is required when creating resources, although\nsome resources may allow a client to request the generation of an appropriate name\nautomatically. Name is primarily intended for creation idempotence and configuration\ndefinition.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names\n+optional",
          "type": "string"
        },
        "namespace": {
          "description": "Namespace defines the space within each name must be unique. An empty namespace is\nequivalent to the \"default\" namespace, but \"default\" is the canonical representation.\nNot all objects are required to be scoped to a namespace - the value of this field for\nthose objects will be empty.\n\nMust be a DNS_LABEL.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/namespaces\n+optional",
          "type": "string"
        },
        "ownerReferences": {
          "items": {
            "$ref": "#/definitions/v1OwnerReference"
          },
          "title": "List of objects depended by this object. If ALL objects in the list have\nbeen deleted, this object will be garbage collected. If this object is managed by a controller,\nthen an entry in this list will point to this controller, with the controller field set to true.\nThere cannot be more than one managing controller.\n+optional\n+patchMergeKey=uid\n+patchStrategy=merge",
          "type": "array"
        },
        "resourceVersion": {
          "description": "An opaque value that represents the internal version of this object that can\nbe used by clients to determine when objects have changed. May be used for optimistic\nconcurrency, change detection, and the watch operation on a resource or set of resources.\nClients must treat these values as opaque and passed unmodified back to the server.\nThey may only be valid for a particular resource or set of resources.\n\nPopulated by the system.\nRead-only.\nValue must be treated as opaque by clients and .\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "SelfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        },
        "uid": {
          "description": "UID is the unique in time and space value for this object. It is typically generated by\nthe server on successful creation of a resource and is not allowed to change on PUT\noperations.\n\nPopulated by the system.\nRead-only.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "apisv1alpha1NodeSpec": {
      "properties": {
        "externalIPType": {
          "type": "string"
        },
        "kubeletExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "nodeDiskSize": {
          "type": "integer"
        },
        "nodeDiskType": {
          "type": "string"
        },
        "sku": {
          "type": "string"
        },
        "spotPriceMax": {
          "type": "number"
        },
        "type": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializer": {
      "description": "Initializer is information about an initializer that has not yet completed.",
      "properties": {
        "name": {
          "description": "name of the process that is responsible for initializing this object.",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializers": {
      "description": "Initializers tracks the progress of initialization.",
      "properties": {
        "pending": {
          "items": {
            "$ref": "#/definitions/v1Initializer"
          },
          "title": "Pending is a list of initializers that must execute in order before this object is visible.\nWhen the last pending initializer is removed, and no failing result is set, the initializers\nstruct will be set to nil and the object is considered as initialized and visible to all\nclients.\n+patchMergeKey=name\n+patchStrategy=merge",
          "type": "array"
        },
        "result": {
          "$ref": "#/definitions/v1Status",
          "description": "If result is set with the Failure field, the object will be persisted to storage and then deleted,\nensuring that other clients can observe the deletion."
        }
      },
      "type": "object"
    },
    "v1ListMeta": {
      "description": "ListMeta describes metadata that synthetic resources must have, including lists and\nvarious status objects. A resource may have only one of {ObjectMeta, ListMeta}.",
      "properties": {
        "continue": {
          "description": "continue may be set if the user set a limit on the number of items returned, and indicates that\nthe server has more data available. The value is opaque and may be used to issue another request\nto the endpoint that served this list to retrieve the next set of available objects. Continuing a\nlist may not be possible if the server configuration has changed or more than a few minutes have\npassed. The resourceVersion field returned when using this continue value will be identical to\nthe value in the first response.",
          "type": "string"
        },
        "resourceVersion": {
          "title": "String that identifies the server's internal version of this object that\ncan be used by clients to determine when objects have changed.\nValue must be treated as opaque by clients and passed unmodified back to the server.\nPopulated by the system.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "selfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1OwnerReference": {
      "description": "OwnerReference contains enough information to let you identify an owning\nobject. Currently, an owning object must be in the same namespace, so there\nis no namespace field.",
      "properties": {
        "apiVersion": {
          "description": "API version of the referent.",
          "type": "string"
        },
        "blockOwnerDeletion": {
          "title": "If true, AND if the owner has the \"foregroundDeletion\" finalizer, then\nthe owner cannot be deleted from the key-value store until this\nreference is removed.\nDefaults to false.\nTo set this field, a user needs \"delete\" permission of the owner,\notherwise 422 (Unprocessable Entity) will be returned.\n+optional",
          "type": "boolean"
        },
        "controller": {
          "title": "If true, this reference points to the managing controller.\n+optional",
          "type": "boolean"
        },
        "kind": {
          "title": "Kind of the referent.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
          "type": "string"
        },
        "name": {
          "title": "Name of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names",
          "type": "string"
        },
        "uid": {
          "title": "UID of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Status": {
      "description": "Status is a return value for calls that don't return other objects.",
      "properties": {
        "code": {
          "title": "Suggested HTTP return code for this status, 0 if not set.\n+optional",
          "type": "integer"
        },
        "details": {
          "$ref": "#/definitions/v1StatusDetails",
          "title": "Extended data associated with the reason.  Each reason may define its\nown extended details. This field is optional and the data returned\nis not guaranteed to conform to any schema except that defined by\nthe reason type.\n+optional"
        },
        "message": {
          "title": "A human-readable description of the status of this operation.\n+optional",
          "type": "string"
        },
        "metadata": {
          "$ref": "#/definitions/v1ListMeta",
          "title": "Standard list metadata.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional"
        },
        "reason": {
          "title": "A machine-readable description of why this operation is in the\n\"Failure\" status. If this value is empty there\nis no information available. A Reason clarifies an HTTP status\ncode but does not override it.\n+optional",
          "type": "string"
        },
        "status": {
          "title": "Status of the operation.\nOne of: \"Success\" or \"Failure\".\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusCause": {
      "description": "StatusCause provides more information about an api.Status failure, including\ncases when multiple errors are encountered.",
      "properties": {
        "field": {
          "description": "The field of the resource that has caused this error, as named by its JSON\nserialization. May include dot and postfix notation for nested attributes.\nArrays are zero-indexed.  Fields may appear more than once in an array of\ncauses due to fields having multiple errors.\nOptional.\n\nExamples:\n  \"name\" - the field \"name\" on the current resource\n  \"items[0].name\" - the field \"name\" on the first array entry in \"items\"\n+optional",
          "type": "string"
        },
        "message": {
          "title": "A human-readable description of the cause of the error.  This field may be\npresented as-is to a reader.\n+optional",
          "type": "string"
        },
        "reason": {
          "title": "A machine-readable description of the cause of the error. If this value is\nempty there is no information available.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusDetails": {
      "description": "StatusDetails is a set of additional properties that MAY be set by the\nserver to provide additional information about a response. The Reason\nfield of a Status object defines what attributes will be set. Clients\nmust ignore fields that do not match the defined type of each attribute,\nand should assume that any attribute may be empty, invalid, or under\ndefined.",
      "properties": {
        "causes": {
          "items": {
            "$ref": "#/definitions/v1StatusCause"
          },
          "title": "The Causes array includes more details associated with the StatusReason\nfailure. Not all StatusReasons may provide detailed causes.\n+optional",
          "type": "array"
        },
        "group": {
          "title": "The group attribute of the resource associated with the status StatusReason.\n+optional",
          "type": "string"
        },
        "kind": {
          "title": "The kind attribute of the resource associated with the status StatusReason.\nOn some operations may differ from the requested resource Kind.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional",
          "type": "string"
        },
        "name": {
          "title": "The name attribute of the resource associated with the status StatusReason\n(when there is a single name which can be described).\n+optional",
          "type": "string"
        },
        "retryAfterSeconds": {
          "title": "If specified, the time in seconds before the operation should be retried. Some errors may indicate\nthe client must take an alternate action - for those errors this field may indicate how long to wait\nbefore taking the alternate action.\n+optional",
          "type": "integer"
        },
        "uid": {
          "title": "UID of the resource.\n(when there is a single resource which can be described).\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Time": {
      "description": "Time is a wrapper around time.Time which supports correct\nmarshaling to YAML and JSON.  Wrappers are provided for many\nof the factory methods that the time package offers.\n\n+protobuf.options.marshal=false\n+protobuf.as=Timestamp\n+protobuf.options.(gogoproto.goproto_stringer)=false",
      "properties": {
        "nanos": {
          "description": "Non-negative fractions of a second at nanosecond resolution. Negative\nsecond values with fractions must still have non-negative nanos values\nthat count forward in time. Must be from 0 to 999,999,999\ninclusive. This field may be limited in precision depending on context.",
          "type": "integer"
        },
        "seconds": {
          "description": "Represents seconds of UTC time since Unix epoch\n1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to\n9999-12-31T23:59:59Z inclusive.",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1NodeGroup": {
      "properties": {
        "metadata": {
          "$ref": "#/definitions/apismetav1ObjectMeta"
        },
        "spec": {
          "$ref": "#/definitions/v1alpha1NodeGroupSpec"
        },
        "status": {
          "$ref": "#/definitions/v1alpha1NodeGroupStatus"
        }
      },
      "type": "object"
    },
    "v1alpha1NodeGroupSpec": {
      "properties": {
        "nodes": {
          "type": "integer"
        },
        "template": {
          "$ref": "#/definitions/v1alpha1NodeTemplateSpec",
          "description": "Template describes the nodes that will be created."
        }
      },
      "type": "object"
    },
    "v1alpha1NodeGroupStatus": {
      "description": "NodeGroupStatus is the most recently observed status of the NodeGroup.",
      "properties": {
        "nodes": {
          "description": "Nodes is the most recently oberved number of nodes.",
          "type": "integer"
        },
        "observedGeneration": {
          "title": "ObservedGeneration reflects the generation of the most recently observed node group.\n+optional",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1NodeTemplateSpec": {
      "properties": {
        "spec": {
          "$ref": "#/definitions/apisv1alpha1NodeSpec",
          "title": "Specification of the desired behavior of the pod.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status\n+optional"
        }
      },
      "title": "PodTemplateSpec describes the data a pod should have when created from a template",
      "type": "object"
    }
  },
  "properties": {
    "nodeGroup": {
      "$ref": "#/definitions/v1alpha1NodeGroup"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	nodeGroupCreateRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "apismetav1ObjectMeta": {
      "description": "ObjectMeta is metadata that all persisted resources must have, which includes all objects\nusers must create.",
      "properties": {
        "annotations": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Annotations is an unstructured key value map stored with a resource that may be\nset by external tools to store and retrieve arbitrary metadata. They are not\nqueryable and should be preserved when modifying objects.\nMore info: http://kubernetes.io/docs/user-guide/annotations\n+optional",
          "type": "object"
        },
        "clusterName": {
          "title": "The name of the cluster which the object belongs to.\nThis is used to distinguish resources with same name and namespace in different clusters.\nThis field is not set anywhere right now and apiserver is going to ignore it if set in create or update request.\n+optional",
          "type": "string"
        },
        "creationTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "CreationTimestamp is a timestamp representing the server time when this object was\ncreated. It is not guaranteed to be set in happens-before order across separate operations.\nClients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system.\nRead-only.\nNull for lists.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "deletionGracePeriodSeconds": {
          "title": "Number of seconds allowed for this object to gracefully terminate before\nit will be removed from the system. Only set when deletionTimestamp is also set.\nMay only be shortened.\nRead-only.\n+optional",
          "type": "integer"
        },
        "deletionTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted. This\nfield is set by the server when a graceful deletion is requested by the user, and is not\ndirectly settable by a client. The resource is expected to be deleted (no longer visible\nfrom resource lists, and not reachable by name) after the time in this field, once the\nfinalizers list is empty. As long as the finalizers list contains items, deletion is blocked.\nOnce the deletionTimestamp is set, this value may not be unset or be set further into the\nfuture, although it may be shortened or the resource may be deleted prior to this time.\nFor example, a user may request that a pod is deleted in 30 seconds. The Kubelet will react\nby sending a graceful termination signal to the containers in the pod. After that 30 seconds,\nthe Kubelet will send a hard termination signal (SIGKILL) to the container and after cleanup,\nremove the pod from the API. In the presence of network partitions, this object may still\nexist after this timestamp, until an administrator or automated process can determine the\nresource is fully terminated.\nIf not set, graceful deletion of the object has not been requested.\n\nPopulated by the system when a graceful deletion is requested.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "finalizers": {
          "items": {
            "type": "string"
          },
          "title": "Must be empty before the object is deleted from the registry. Each entry\nis an identifier for the responsible component that will remove the entry\nfrom the list. If the deletionTimestamp of the object is non-nil, entries\nin this list can only be removed.\n+optional\n+patchStrategy=merge",
          "type": "array"
        },
        "generateName": {
          "description": "GenerateName is an optional prefix, used by the server, to generate a unique\nname ONLY IF the Name field has not been provided.\nIf this field is used, the name returned to the client will be different\nthan the name passed. This value will also be combined with a unique suffix.\nThe provided value has the same validation rules as the Name field,\nand may be truncated by the length of the suffix required to make the value\nunique on the server.\n\nIf this field is specified and the generated name exists, the server will\nNOT return a 409 - instead, it will either return 201 Created or 500 with Reason\nServerTimeout indicating a unique name could not be found in the time allotted, and the client\nshould retry (optionally after the time indicated in the Retry-After header).\n\nApplied only if Name is not specified.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency\n+optional",
          "type": "string"
        },
        "generation": {
          "title": "A sequence number representing a specific generation of the desired state.\nPopulated by the system. Read-only.\n+optional",
          "type": "integer"
        },
        "initializers": {
          "$ref": "#/definitions/v1Initializers",
          "description": "An initializer is a controller which enforces some system invariant at object creation time.\nThis field is a list of initializers that have not yet acted on this object. If nil or empty,\nthis object has been completely initialized. Otherwise, the object is considered uninitialized\nand is hidden (in list/watch and get calls) from clients that haven't explicitly asked to\nobserve uninitialized objects.\n\nWhen an object is created, the system will populate this list with the current set of initializers.\nOnly privileged users may set or modify this list. Once it is empty, it may not be modified further\nby any user."
        },
        "labels": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Map of string keys and values that can be used to organize and categorize\n(scope and select) objects. May match selectors of replication controllers\nand services.\nMore info: http://kubernetes.io/docs/user-guide/labels\n+optional",
          "type": "object"
        },
        "name": {
          "title": "Name must be unique within a namespace. Is required when creating resources, although\nsome resources may allow a client to request the generation of an appropriate name\nautomatically. Name is primarily intended for creation idempotence and configuration\ndefinition.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names\n+optional",
          "type": "string"
        },
        "namespace": {
          "description": "Namespace defines the space within each name must be unique. An empty namespace is\nequivalent to the \"default\" namespace, but \"default\" is the canonical representation.\nNot all objects are required to be scoped to a namespace - the value of this field for\nthose objects will be empty.\n\nMust be a DNS_LABEL.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/namespaces\n+optional",
          "type": "string"
        },
        "ownerReferences": {
          "items": {
            "$ref": "#/definitions/v1OwnerReference"
          },
          "title": "List of objects depended by this object. If ALL objects in the list have\nbeen deleted, this object will be garbage collected. If this object is managed by a controller,\nthen an entry in this list will point to this controller, with the controller field set to true.\nThere cannot be more than one managing controller.\n+optional\n+patchMergeKey=uid\n+patchStrategy=merge",
          "type": "array"
        },
        "resourceVersion": {
          "description": "An opaque value that represents the internal version of this object that can\nbe used by clients to determine when objects have changed. May be used for optimistic\nconcurrency, change detection, and the watch operation on a resource or set of resources.\nClients must treat these values as opaque and passed unmodified back to the server.\nThey may only be valid for a particular resource or set of resources.\n\nPopulated by the system.\nRead-only.\nValue must be treated as opaque by clients and .\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "SelfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        },
        "uid": {
          "description": "UID is the unique in time and space value for this object. It is typically generated by\nthe server on successful creation of a resource and is not allowed to change on PUT\noperations.\n\nPopulated by the system.\nRead-only.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "apisv1alpha1NodeSpec": {
      "properties": {
        "externalIPType": {
          "type": "string"
        },
        "kubeletExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "nodeDiskSize": {
          "type": "integer"
        },
        "nodeDiskType": {
          "type": "string"
        },
        "sku": {
          "type": "string"
        },
        "spotPriceMax": {
          "type": "number"
        },
        "type": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializer": {
      "description": "Initializer is information about an initializer that has not yet completed.",
      "properties": {
        "name": {
          "description": "name of the process that is responsible for initializing this object.",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializers": {
      "description": "Initializers tracks the progress of initialization.",
      "properties": {
        "pending": {
          "items": {
            "$ref": "#/definitions/v1Initializer"
          },
          "title": "Pending is a list of initializers that must execute in order before this object is visible.\nWhen the last pending initializer is removed, and no failing result is set, the initializers\nstruct will be set to nil and the object is considered as initialized and visible to all\nclients.\n+patchMergeKey=name\n+patchStrategy=merge",
          "type": "array"
        },
        "result": {
          "$ref": "#/definitions/v1Status",
          "description": "If result is set with the Failure field, the object will be persisted to storage and then deleted,\nensuring that other clients can observe the deletion."
        }
      },
      "type": "object"
    },
    "v1ListMeta": {
      "description": "ListMeta describes metadata that synthetic resources must have, including lists and\nvarious status objects. A resource may have only one of {ObjectMeta, ListMeta}.",
      "properties": {
        "continue": {
          "description": "continue may be set if the user set a limit on the number of items returned, and indicates that\nthe server has more data available. The value is opaque and may be used to issue another request\nto the endpoint that served this list to retrieve the next set of available objects. Continuing a\nlist may not be possible if the server configuration has changed or more than a few minutes have\npassed. The resourceVersion field returned when using this continue value will be identical to\nthe value in the first response.",
          "type": "string"
        },
        "resourceVersion": {
          "title": "String that identifies the server's internal version of this object that\ncan be used by clients to determine when objects have changed.\nValue must be treated as opaque by clients and passed unmodified back to the server.\nPopulated by the system.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "selfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1OwnerReference": {
      "description": "OwnerReference contains enough information to let you identify an owning\nobject. Currently, an owning object must be in the same namespace, so there\nis no namespace field.",
      "properties": {
        "apiVersion": {
          "description": "API version of the referent.",
          "type": "string"
        },
        "blockOwnerDeletion": {
          "title": "If true, AND if the owner has the \"foregroundDeletion\" finalizer, then\nthe owner cannot be deleted from the key-value store until this\nreference is removed.\nDefaults to false.\nTo set this field, a user needs \"delete\" permission of the owner,\notherwise 422 (Unprocessable Entity) will be returned.\n+optional",
          "type": "boolean"
        },
        "controller": {
          "title": "If true, this reference points to the managing controller.\n+optional",
          "type": "boolean"
        },
        "kind": {
          "title": "Kind of the referent.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
          "type": "string"
        },
        "name": {
          "title": "Name of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names",
          "type": "string"
        },
        "uid": {
          "title": "UID of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Status": {
      "description": "Status is a return value for calls that don't return other objects.",
      "properties": {
        "code": {
          "title": "Suggested HTTP return code for this status, 0 if not set.\n+optional",
          "type": "integer"
        },
        "details": {
          "$ref": "#/definitions/v1StatusDetails",
          "title": "Extended data associated with the reason.  Each reason may define its\nown extended details. This field is optional and the data returned\nis not guaranteed to conform to any schema except that defined by\nthe reason type.\n+optional"
        },
        "message": {
          "title": "A human-readable description of the status of this operation.\n+optional",
          "type": "string"
        },
        "metadata": {
          "$ref": "#/definitions/v1ListMeta",
          "title": "Standard list metadata.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional"
        },
        "reason": {
          "title": "A machine-readable description of why this operation is in the\n\"Failure\" status. If this value is empty there\nis no information available. A Reason clarifies an HTTP status\ncode but does not override it.\n+optional",
          "type": "string"
        },
        "status": {
          "title": "Status of the operation.\nOne of: \"Success\" or \"Failure\".\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusCause": {
      "description": "StatusCause provides more information about an api.Status failure, including\ncases when multiple errors are encountered.",
      "properties": {
        "field": {
          "description": "The field of the resource that has caused this error, as named by its JSON\nserialization. May include dot and postfix notation for nested attributes.\nArrays are zero-indexed.  Fields may appear more than once in an array of\ncauses due to fields having multiple errors.\nOptional.\n\nExamples:\n  \"name\" - the field \"name\" on the current resource\n  \"items[0].name\" - the field \"name\" on the first array entry in \"items\"\n+optional",
          "type": "string"
        },
        "message": {
          "title": "A human-readable description of the cause of the error.  This field may be\npresented as-is to a reader.\n+optional",
          "type": "string"
        },
        "reason": {
          "title": "A machine-readable description of the cause of the error. If this value is\nempty there is no information available.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusDetails": {
      "description": "StatusDetails is a set of additional properties that MAY be set by the\nserver to provide additional information about a response. The Reason\nfield of a Status object defines what attributes will be set. Clients\nmust ignore fields that do not match the defined type of each attribute,\nand should assume that any attribute may be empty, invalid, or under\ndefined.",
      "properties": {
        "causes": {
          "items": {
            "$ref": "#/definitions/v1StatusCause"
          },
          "title": "The Causes array includes more details associated with the StatusReason\nfailure. Not all StatusReasons may provide detailed causes.\n+optional",
          "type": "array"
        },
        "group": {
          "title": "The group attribute of the resource associated with the status StatusReason.\n+optional",
          "type": "string"
        },
        "kind": {
          "title": "The kind attribute of the resource associated with the status StatusReason.\nOn some operations may differ from the requested resource Kind.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional",
          "type": "string"
        },
        "name": {
          "title": "The name attribute of the resource associated with the status StatusReason\n(when there is a single name which can be described).\n+optional",
          "type": "string"
        },
        "retryAfterSeconds": {
          "title": "If specified, the time in seconds before the operation should be retried. Some errors may indicate\nthe client must take an alternate action - for those errors this field may indicate how long to wait\nbefore taking the alternate action.\n+optional",
          "type": "integer"
        },
        "uid": {
          "title": "UID of the resource.\n(when there is a single resource which can be described).\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Time": {
      "description": "Time is a wrapper around time.Time which supports correct\nmarshaling to YAML and JSON.  Wrappers are provided for many\nof the factory methods that the time package offers.\n\n+protobuf.options.marshal=false\n+protobuf.as=Timestamp\n+protobuf.options.(gogoproto.goproto_stringer)=false",
      "properties": {
        "nanos": {
          "description": "Non-negative fractions of a second at nanosecond resolution. Negative\nsecond values with fractions must still have non-negative nanos values\nthat count forward in time. Must be from 0 to 999,999,999\ninclusive. This field may be limited in precision depending on context.",
          "type": "integer"
        },
        "seconds": {
          "description": "Represents seconds of UTC time since Unix epoch\n1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to\n9999-12-31T23:59:59Z inclusive.",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1NodeGroup": {
      "properties": {
        "metadata": {
          "$ref": "#/definitions/apismetav1ObjectMeta"
        },
        "spec": {
          "$ref": "#/definitions/v1alpha1NodeGroupSpec"
        },
        "status": {
          "$ref": "#/definitions/v1alpha1NodeGroupStatus"
        }
      },
      "type": "object"
    },
    "v1alpha1NodeGroupSpec": {
      "properties": {
        "nodes": {
          "type": "integer"
        },
        "template": {
          "$ref": "#/definitions/v1alpha1NodeTemplateSpec",
          "description": "Template describes the nodes that will be created."
        }
      },
      "type": "object"
    },
    "v1alpha1NodeGroupStatus": {
      "description": "NodeGroupStatus is the most recently observed status of the NodeGroup.",
      "properties": {
        "nodes": {
          "description": "Nodes is the most recently oberved number of nodes.",
          "type": "integer"
        },
        "observedGeneration": {
          "title": "ObservedGeneration reflects the generation of the most recently observed node group.\n+optional",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1NodeTemplateSpec": {
      "properties": {
        "spec": {
          "$ref": "#/definitions/apisv1alpha1NodeSpec",
          "title": "Specification of the desired behavior of the pod.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status\n+optional"
        }
      },
      "title": "PodTemplateSpec describes the data a pod should have when created from a template",
      "type": "object"
    }
  },
  "properties": {
    "nodeGroup": {
      "$ref": "#/definitions/v1alpha1NodeGroup"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	credentialUpdateRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "apismetav1ObjectMeta": {
      "description": "ObjectMeta is metadata that all persisted resources must have, which includes all objects\nusers must create.",
      "properties": {
        "annotations": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Annotations is an unstructured key value map stored with a resource that may be\nset by external tools to store and retrieve arbitrary metadata. They are not\nqueryable and should be preserved when modifying objects.\nMore info: http://kubernetes.io/docs/user-guide/annotations\n+optional",
          "type": "object"
        },
        "clusterName": {
          "title": "The name of the cluster which the object belongs to.\nThis is used to distinguish resources with same name and namespace in different clusters.\nThis field is not set anywhere right now and apiserver is going to ignore it if set in create or update request.\n+optional",
          "type": "string"
        },
        "creationTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "CreationTimestamp is a timestamp representing the server time when this object was\ncreated. It is not guaranteed to be set in happens-before order across separate operations.\nClients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system.\nRead-only.\nNull for lists.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "deletionGracePeriodSeconds": {
          "title": "Number of seconds allowed for this object to gracefully terminate before\nit will be removed from the system. Only set when deletionTimestamp is also set.\nMay only be shortened.\nRead-only.\n+optional",
          "type": "integer"
        },
        "deletionTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted. This\nfield is set by the server when a graceful deletion is requested by the user, and is not\ndirectly settable by a client. The resource is expected to be deleted (no longer visible\nfrom resource lists, and not reachable by name) after the time in this field, once the\nfinalizers list is empty. As long as the finalizers list contains items, deletion is blocked.\nOnce the deletionTimestamp is set, this value may not be unset or be set further into the\nfuture, although it may be shortened or the resource may be deleted prior to this time.\nFor example, a user may request that a pod is deleted in 30 seconds. The Kubelet will react\nby sending a graceful termination signal to the containers in the pod. After that 30 seconds,\nthe Kubelet will send a hard termination signal (SIGKILL) to the container and after cleanup,\nremove the pod from the API. In the presence of network partitions, this object may still\nexist after this timestamp, until an administrator or automated process can determine the\nresource is fully terminated.\nIf not set, graceful deletion of the object has not been requested.\n\nPopulated by the system when a graceful deletion is requested.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "finalizers": {
          "items": {
            "type": "string"
          },
          "title": "Must be empty before the object is deleted from the registry. Each entry\nis an identifier for the responsible component that will remove the entry\nfrom the list. If the deletionTimestamp of the object is non-nil, entries\nin this list can only be removed.\n+optional\n+patchStrategy=merge",
          "type": "array"
        },
        "generateName": {
          "description": "GenerateName is an optional prefix, used by the server, to generate a unique\nname ONLY IF the Name field has not been provided.\nIf this field is used, the name returned to the client will be different\nthan the name passed. This value will also be combined with a unique suffix.\nThe provided value has the same validation rules as the Name field,\nand may be truncated by the length of the suffix required to make the value\nunique on the server.\n\nIf this field is specified and the generated name exists, the server will\nNOT return a 409 - instead, it will either return 201 Created or 500 with Reason\nServerTimeout indicating a unique name could not be found in the time allotted, and the client\nshould retry (optionally after the time indicated in the Retry-After header).\n\nApplied only if Name is not specified.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency\n+optional",
          "type": "string"
        },
        "generation": {
          "title": "A sequence number representing a specific generation of the desired state.\nPopulated by the system. Read-only.\n+optional",
          "type": "integer"
        },
        "initializers": {
          "$ref": "#/definitions/v1Initializers",
          "description": "An initializer is a controller which enforces some system invariant at object creation time.\nThis field is a list of initializers that have not yet acted on this object. If nil or empty,\nthis object has been completely initialized. Otherwise, the object is considered uninitialized\nand is hidden (in list/watch and get calls) from clients that haven't explicitly asked to\nobserve uninitialized objects.\n\nWhen an object is created, the system will populate this list with the current set of initializers.\nOnly privileged users may set or modify this list. Once it is empty, it may not be modified further\nby any user."
        },
        "labels": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Map of string keys and values that can be used to organize and categorize\n(scope and select) objects. May match selectors of replication controllers\nand services.\nMore info: http://kubernetes.io/docs/user-guide/labels\n+optional",
          "type": "object"
        },
        "name": {
          "title": "Name must be unique within a namespace. Is required when creating resources, although\nsome resources may allow a client to request the generation of an appropriate name\nautomatically. Name is primarily intended for creation idempotence and configuration\ndefinition.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names\n+optional",
          "type": "string"
        },
        "namespace": {
          "description": "Namespace defines the space within each name must be unique. An empty namespace is\nequivalent to the \"default\" namespace, but \"default\" is the canonical representation.\nNot all objects are required to be scoped to a namespace - the value of this field for\nthose objects will be empty.\n\nMust be a DNS_LABEL.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/namespaces\n+optional",
          "type": "string"
        },
        "ownerReferences": {
          "items": {
            "$ref": "#/definitions/v1OwnerReference"
          },
          "title": "List of objects depended by this object. If ALL objects in the list have\nbeen deleted, this object will be garbage collected. If this object is managed by a controller,\nthen an entry in this list will point to this controller, with the controller field set to true.\nThere cannot be more than one managing controller.\n+optional\n+patchMergeKey=uid\n+patchStrategy=merge",
          "type": "array"
        },
        "resourceVersion": {
          "description": "An opaque value that represents the internal version of this object that can\nbe used by clients to determine when objects have changed. May be used for optimistic\nconcurrency, change detection, and the watch operation on a resource or set of resources.\nClients must treat these values as opaque and passed unmodified back to the server.\nThey may only be valid for a particular resource or set of resources.\n\nPopulated by the system.\nRead-only.\nValue must be treated as opaque by clients and .\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "SelfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        },
        "uid": {
          "description": "UID is the unique in time and space value for this object. It is typically generated by\nthe server on successful creation of a resource and is not allowed to change on PUT\noperations.\n\nPopulated by the system.\nRead-only.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializer": {
      "description": "Initializer is information about an initializer that has not yet completed.",
      "properties": {
        "name": {
          "description": "name of the process that is responsible for initializing this object.",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializers": {
      "description": "Initializers tracks the progress of initialization.",
      "properties": {
        "pending": {
          "items": {
            "$ref": "#/definitions/v1Initializer"
          },
          "title": "Pending is a list of initializers that must execute in order before this object is visible.\nWhen the last pending initializer is removed, and no failing result is set, the initializers\nstruct will be set to nil and the object is considered as initialized and visible to all\nclients.\n+patchMergeKey=name\n+patchStrategy=merge",
          "type": "array"
        },
        "result": {
          "$ref": "#/definitions/v1Status",
          "description": "If result is set with the Failure field, the object will be persisted to storage and then deleted,\nensuring that other clients can observe the deletion."
        }
      },
      "type": "object"
    },
    "v1ListMeta": {
      "description": "ListMeta describes metadata that synthetic resources must have, including lists and\nvarious status objects. A resource may have only one of {ObjectMeta, ListMeta}.",
      "properties": {
        "continue": {
          "description": "continue may be set if the user set a limit on the number of items returned, and indicates that\nthe server has more data available. The value is opaque and may be used to issue another request\nto the endpoint that served this list to retrieve the next set of available objects. Continuing a\nlist may not be possible if the server configuration has changed or more than a few minutes have\npassed. The resourceVersion field returned when using this continue value will be identical to\nthe value in the first response.",
          "type": "string"
        },
        "resourceVersion": {
          "title": "String that identifies the server's internal version of this object that\ncan be used by clients to determine when objects have changed.\nValue must be treated as opaque by clients and passed unmodified back to the server.\nPopulated by the system.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "selfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1OwnerReference": {
      "description": "OwnerReference contains enough information to let you identify an owning\nobject. Currently, an owning object must be in the same namespace, so there\nis no namespace field.",
      "properties": {
        "apiVersion": {
          "description": "API version of the referent.",
          "type": "string"
        },
        "blockOwnerDeletion": {
          "title": "If true, AND if the owner has the \"foregroundDeletion\" finalizer, then\nthe owner cannot be deleted from the key-value store until this\nreference is removed.\nDefaults to false.\nTo set this field, a user needs \"delete\" permission of the owner,\notherwise 422 (Unprocessable Entity) will be returned.\n+optional",
          "type": "boolean"
        },
        "controller": {
          "title": "If true, this reference points to the managing controller.\n+optional",
          "type": "boolean"
        },
        "kind": {
          "title": "Kind of the referent.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
          "type": "string"
        },
        "name": {
          "title": "Name of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names",
          "type": "string"
        },
        "uid": {
          "title": "UID of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Status": {
      "description": "Status is a return value for calls that don't return other objects.",
      "properties": {
        "code": {
          "title": "Suggested HTTP return code for this status, 0 if not set.\n+optional",
          "type": "integer"
        },
        "details": {
          "$ref": "#/definitions/v1StatusDetails",
          "title": "Extended data associated with the reason.  Each reason may define its\nown extended details. This field is optional and the data returned\nis not guaranteed to conform to any schema except that defined by\nthe reason type.\n+optional"
        },
        "message": {
          "title": "A human-readable description of the status of this operation.\n+optional",
          "type": "string"
        },
        "metadata": {
          "$ref": "#/definitions/v1ListMeta",
          "title": "Standard list metadata.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional"
        },
        "reason": {
          "title": "A machine-readable description of why this operation is in the\n\"Failure\" status. If this value is empty there\nis no information available. A Reason clarifies an HTTP status\ncode but does not override it.\n+optional",
          "type": "string"
        },
        "status": {
          "title": "Status of the operation.\nOne of: \"Success\" or \"Failure\".\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusCause": {
      "description": "StatusCause provides more information about an api.Status failure, including\ncases when multiple errors are encountered.",
      "properties": {
        "field": {
          "description": "The field of the resource that has caused this error, as named by its JSON\nserialization. May include dot and postfix notation for nested attributes.\nArrays are zero-indexed.  Fields may appear more than once in an array of\ncauses due to fields having multiple errors.\nOptional.\n\nExamples:\n  \"name\" - the field \"name\" on the current resource\n  \"items[0].name\" - the field \"name\" on the first array entry in \"items\"\n+optional",
          "type": "string"
        },
        "message": {
          "title": "A human-readable description of the cause of the error.  This field may be\npresented as-is to a reader.\n+optional",
          "type": "string"
        },
        "reason": {
          "title": "A machine-readable description of the cause of the error. If this value is\nempty there is no information available.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusDetails": {
      "description": "StatusDetails is a set of additional properties that MAY be set by the\nserver to provide additional information about a response. The Reason\nfield of a Status object defines what attributes will be set. Clients\nmust ignore fields that do not match the defined type of each attribute,\nand should assume that any attribute may be empty, invalid, or under\ndefined.",
      "properties": {
        "causes": {
          "items": {
            "$ref": "#/definitions/v1StatusCause"
          },
          "title": "The Causes array includes more details associated with the StatusReason\nfailure. Not all StatusReasons may provide detailed causes.\n+optional",
          "type": "array"
        },
        "group": {
          "title": "The group attribute of the resource associated with the status StatusReason.\n+optional",
          "type": "string"
        },
        "kind": {
          "title": "The kind attribute of the resource associated with the status StatusReason.\nOn some operations may differ from the requested resource Kind.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional",
          "type": "string"
        },
        "name": {
          "title": "The name attribute of the resource associated with the status StatusReason\n(when there is a single name which can be described).\n+optional",
          "type": "string"
        },
        "retryAfterSeconds": {
          "title": "If specified, the time in seconds before the operation should be retried. Some errors may indicate\nthe client must take an alternate action - for those errors this field may indicate how long to wait\nbefore taking the alternate action.\n+optional",
          "type": "integer"
        },
        "uid": {
          "title": "UID of the resource.\n(when there is a single resource which can be described).\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Time": {
      "description": "Time is a wrapper around time.Time which supports correct\nmarshaling to YAML and JSON.  Wrappers are provided for many\nof the factory methods that the time package offers.\n\n+protobuf.options.marshal=false\n+protobuf.as=Timestamp\n+protobuf.options.(gogoproto.goproto_stringer)=false",
      "properties": {
        "nanos": {
          "description": "Non-negative fractions of a second at nanosecond resolution. Negative\nsecond values with fractions must still have non-negative nanos values\nthat count forward in time. Must be from 0 to 999,999,999\ninclusive. This field may be limited in precision depending on context.",
          "type": "integer"
        },
        "seconds": {
          "description": "Represents seconds of UTC time since Unix epoch\n1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to\n9999-12-31T23:59:59Z inclusive.",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1Credential": {
      "properties": {
        "metadata": {
          "$ref": "#/definitions/apismetav1ObjectMeta"
        },
        "spec": {
          "$ref": "#/definitions/v1alpha1CredentialSpec"
        }
      },
      "type": "object"
    },
    "v1alpha1CredentialSpec": {
      "properties": {
        "data": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "provider": {
          "type": "string"
        }
      },
      "type": "object"
    }
  },
  "properties": {
    "credential": {
      "$ref": "#/definitions/v1alpha1Credential"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	clusterCreateRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "apismetav1ObjectMeta": {
      "description": "ObjectMeta is metadata that all persisted resources must have, which includes all objects\nusers must create.",
      "properties": {
        "annotations": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Annotations is an unstructured key value map stored with a resource that may be\nset by external tools to store and retrieve arbitrary metadata. They are not\nqueryable and should be preserved when modifying objects.\nMore info: http://kubernetes.io/docs/user-guide/annotations\n+optional",
          "type": "object"
        },
        "clusterName": {
          "title": "The name of the cluster which the object belongs to.\nThis is used to distinguish resources with same name and namespace in different clusters.\nThis field is not set anywhere right now and apiserver is going to ignore it if set in create or update request.\n+optional",
          "type": "string"
        },
        "creationTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "CreationTimestamp is a timestamp representing the server time when this object was\ncreated. It is not guaranteed to be set in happens-before order across separate operations.\nClients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system.\nRead-only.\nNull for lists.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "deletionGracePeriodSeconds": {
          "title": "Number of seconds allowed for this object to gracefully terminate before\nit will be removed from the system. Only set when deletionTimestamp is also set.\nMay only be shortened.\nRead-only.\n+optional",
          "type": "integer"
        },
        "deletionTimestamp": {
          "$ref": "#/definitions/v1Time",
          "description": "DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted. This\nfield is set by the server when a graceful deletion is requested by the user, and is not\ndirectly settable by a client. The resource is expected to be deleted (no longer visible\nfrom resource lists, and not reachable by name) after the time in this field, once the\nfinalizers list is empty. As long as the finalizers list contains items, deletion is blocked.\nOnce the deletionTimestamp is set, this value may not be unset or be set further into the\nfuture, although it may be shortened or the resource may be deleted prior to this time.\nFor example, a user may request that a pod is deleted in 30 seconds. The Kubelet will react\nby sending a graceful termination signal to the containers in the pod. After that 30 seconds,\nthe Kubelet will send a hard termination signal (SIGKILL) to the container and after cleanup,\nremove the pod from the API. In the presence of network partitions, this object may still\nexist after this timestamp, until an administrator or automated process can determine the\nresource is fully terminated.\nIf not set, graceful deletion of the object has not been requested.\n\nPopulated by the system when a graceful deletion is requested.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata\n+optional"
        },
        "finalizers": {
          "items": {
            "type": "string"
          },
          "title": "Must be empty before the object is deleted from the registry. Each entry\nis an identifier for the responsible component that will remove the entry\nfrom the list. If the deletionTimestamp of the object is non-nil, entries\nin this list can only be removed.\n+optional\n+patchStrategy=merge",
          "type": "array"
        },
        "generateName": {
          "description": "GenerateName is an optional prefix, used by the server, to generate a unique\nname ONLY IF the Name field has not been provided.\nIf this field is used, the name returned to the client will be different\nthan the name passed. This value will also be combined with a unique suffix.\nThe provided value has the same validation rules as the Name field,\nand may be truncated by the length of the suffix required to make the value\nunique on the server.\n\nIf this field is specified and the generated name exists, the server will\nNOT return a 409 - instead, it will either return 201 Created or 500 with Reason\nServerTimeout indicating a unique name could not be found in the time allotted, and the client\nshould retry (optionally after the time indicated in the Retry-After header).\n\nApplied only if Name is not specified.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency\n+optional",
          "type": "string"
        },
        "generation": {
          "title": "A sequence number representing a specific generation of the desired state.\nPopulated by the system. Read-only.\n+optional",
          "type": "integer"
        },
        "initializers": {
          "$ref": "#/definitions/v1Initializers",
          "description": "An initializer is a controller which enforces some system invariant at object creation time.\nThis field is a list of initializers that have not yet acted on this object. If nil or empty,\nthis object has been completely initialized. Otherwise, the object is considered uninitialized\nand is hidden (in list/watch and get calls) from clients that haven't explicitly asked to\nobserve uninitialized objects.\n\nWhen an object is created, the system will populate this list with the current set of initializers.\nOnly privileged users may set or modify this list. Once it is empty, it may not be modified further\nby any user."
        },
        "labels": {
          "additionalProperties": {
            "type": "string"
          },
          "title": "Map of string keys and values that can be used to organize and categorize\n(scope and select) objects. May match selectors of replication controllers\nand services.\nMore info: http://kubernetes.io/docs/user-guide/labels\n+optional",
          "type": "object"
        },
        "name": {
          "title": "Name must be unique within a namespace. Is required when creating resources, although\nsome resources may allow a client to request the generation of an appropriate name\nautomatically. Name is primarily intended for creation idempotence and configuration\ndefinition.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names\n+optional",
          "type": "string"
        },
        "namespace": {
          "description": "Namespace defines the space within each name must be unique. An empty namespace is\nequivalent to the \"default\" namespace, but \"default\" is the canonical representation.\nNot all objects are required to be scoped to a namespace - the value of this field for\nthose objects will be empty.\n\nMust be a DNS_LABEL.\nCannot be updated.\nMore info: http://kubernetes.io/docs/user-guide/namespaces\n+optional",
          "type": "string"
        },
        "ownerReferences": {
          "items": {
            "$ref": "#/definitions/v1OwnerReference"
          },
          "title": "List of objects depended by this object. If ALL objects in the list have\nbeen deleted, this object will be garbage collected. If this object is managed by a controller,\nthen an entry in this list will point to this controller, with the controller field set to true.\nThere cannot be more than one managing controller.\n+optional\n+patchMergeKey=uid\n+patchStrategy=merge",
          "type": "array"
        },
        "resourceVersion": {
          "description": "An opaque value that represents the internal version of this object that can\nbe used by clients to determine when objects have changed. May be used for optimistic\nconcurrency, change detection, and the watch operation on a resource or set of resources.\nClients must treat these values as opaque and passed unmodified back to the server.\nThey may only be valid for a particular resource or set of resources.\n\nPopulated by the system.\nRead-only.\nValue must be treated as opaque by clients and .\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "SelfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        },
        "uid": {
          "description": "UID is the unique in time and space value for this object. It is typically generated by\nthe server on successful creation of a resource and is not allowed to change on PUT\noperations.\n\nPopulated by the system.\nRead-only.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializer": {
      "description": "Initializer is information about an initializer that has not yet completed.",
      "properties": {
        "name": {
          "description": "name of the process that is responsible for initializing this object.",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Initializers": {
      "description": "Initializers tracks the progress of initialization.",
      "properties": {
        "pending": {
          "items": {
            "$ref": "#/definitions/v1Initializer"
          },
          "title": "Pending is a list of initializers that must execute in order before this object is visible.\nWhen the last pending initializer is removed, and no failing result is set, the initializers\nstruct will be set to nil and the object is considered as initialized and visible to all\nclients.\n+patchMergeKey=name\n+patchStrategy=merge",
          "type": "array"
        },
        "result": {
          "$ref": "#/definitions/v1Status",
          "description": "If result is set with the Failure field, the object will be persisted to storage and then deleted,\nensuring that other clients can observe the deletion."
        }
      },
      "type": "object"
    },
    "v1ListMeta": {
      "description": "ListMeta describes metadata that synthetic resources must have, including lists and\nvarious status objects. A resource may have only one of {ObjectMeta, ListMeta}.",
      "properties": {
        "continue": {
          "description": "continue may be set if the user set a limit on the number of items returned, and indicates that\nthe server has more data available. The value is opaque and may be used to issue another request\nto the endpoint that served this list to retrieve the next set of available objects. Continuing a\nlist may not be possible if the server configuration has changed or more than a few minutes have\npassed. The resourceVersion field returned when using this continue value will be identical to\nthe value in the first response.",
          "type": "string"
        },
        "resourceVersion": {
          "title": "String that identifies the server's internal version of this object that\ncan be used by clients to determine when objects have changed.\nValue must be treated as opaque by clients and passed unmodified back to the server.\nPopulated by the system.\nRead-only.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency\n+optional",
          "type": "string"
        },
        "selfLink": {
          "title": "selfLink is a URL representing this object.\nPopulated by the system.\nRead-only.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1NodeAddress": {
      "description": "NodeAddress contains information for the node's address.",
      "properties": {
        "address": {
          "description": "The node address.",
          "type": "string"
        },
        "type": {
          "description": "Node address type, one of Hostname, ExternalIP or InternalIP.",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1OwnerReference": {
      "description": "OwnerReference contains enough information to let you identify an owning\nobject. Currently, an owning object must be in the same namespace, so there\nis no namespace field.",
      "properties": {
        "apiVersion": {
          "description": "API version of the referent.",
          "type": "string"
        },
        "blockOwnerDeletion": {
          "title": "If true, AND if the owner has the \"foregroundDeletion\" finalizer, then\nthe owner cannot be deleted from the key-value store until this\nreference is removed.\nDefaults to false.\nTo set this field, a user needs \"delete\" permission of the owner,\notherwise 422 (Unprocessable Entity) will be returned.\n+optional",
          "type": "boolean"
        },
        "controller": {
          "title": "If true, this reference points to the managing controller.\n+optional",
          "type": "boolean"
        },
        "kind": {
          "title": "Kind of the referent.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
          "type": "string"
        },
        "name": {
          "title": "Name of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#names",
          "type": "string"
        },
        "uid": {
          "title": "UID of the referent.\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Status": {
      "description": "Status is a return value for calls that don't return other objects.",
      "properties": {
        "code": {
          "title": "Suggested HTTP return code for this status, 0 if not set.\n+optional",
          "type": "integer"
        },
        "details": {
          "$ref": "#/definitions/v1StatusDetails",
          "title": "Extended data associated with the reason.  Each reason may define its\nown extended details. This field is optional and the data returned\nis not guaranteed to conform to any schema except that defined by\nthe reason type.\n+optional"
        },
        "message": {
          "title": "A human-readable description of the status of this operation.\n+optional",
          "type": "string"
        },
        "metadata": {
          "$ref": "#/definitions/v1ListMeta",
          "title": "Standard list metadata.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional"
        },
        "reason": {
          "title": "A machine-readable description of why this operation is in the\n\"Failure\" status. If this value is empty there\nis no information available. A Reason clarifies an HTTP status\ncode but does not override it.\n+optional",
          "type": "string"
        },
        "status": {
          "title": "Status of the operation.\nOne of: \"Success\" or \"Failure\".\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusCause": {
      "description": "StatusCause provides more information about an api.Status failure, including\ncases when multiple errors are encountered.",
      "properties": {
        "field": {
          "description": "The field of the resource that has caused this error, as named by its JSON\nserialization. May include dot and postfix notation for nested attributes.\nArrays are zero-indexed.  Fields may appear more than once in an array of\ncauses due to fields having multiple errors.\nOptional.\n\nExamples:\n  \"name\" - the field \"name\" on the current resource\n  \"items[0].name\" - the field \"name\" on the first array entry in \"items\"\n+optional",
          "type": "string"
        },
        "message": {
          "title": "A human-readable description of the cause of the error.  This field may be\npresented as-is to a reader.\n+optional",
          "type": "string"
        },
        "reason": {
          "title": "A machine-readable description of the cause of the error. If this value is\nempty there is no information available.\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1StatusDetails": {
      "description": "StatusDetails is a set of additional properties that MAY be set by the\nserver to provide additional information about a response. The Reason\nfield of a Status object defines what attributes will be set. Clients\nmust ignore fields that do not match the defined type of each attribute,\nand should assume that any attribute may be empty, invalid, or under\ndefined.",
      "properties": {
        "causes": {
          "items": {
            "$ref": "#/definitions/v1StatusCause"
          },
          "title": "The Causes array includes more details associated with the StatusReason\nfailure. Not all StatusReasons may provide detailed causes.\n+optional",
          "type": "array"
        },
        "group": {
          "title": "The group attribute of the resource associated with the status StatusReason.\n+optional",
          "type": "string"
        },
        "kind": {
          "title": "The kind attribute of the resource associated with the status StatusReason.\nOn some operations may differ from the requested resource Kind.\nMore info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds\n+optional",
          "type": "string"
        },
        "name": {
          "title": "The name attribute of the resource associated with the status StatusReason\n(when there is a single name which can be described).\n+optional",
          "type": "string"
        },
        "retryAfterSeconds": {
          "title": "If specified, the time in seconds before the operation should be retried. Some errors may indicate\nthe client must take an alternate action - for those errors this field may indicate how long to wait\nbefore taking the alternate action.\n+optional",
          "type": "integer"
        },
        "uid": {
          "title": "UID of the resource.\n(when there is a single resource which can be described).\nMore info: http://kubernetes.io/docs/user-guide/identifiers#uids\n+optional",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1Time": {
      "description": "Time is a wrapper around time.Time which supports correct\nmarshaling to YAML and JSON.  Wrappers are provided for many\nof the factory methods that the time package offers.\n\n+protobuf.options.marshal=false\n+protobuf.as=Timestamp\n+protobuf.options.(gogoproto.goproto_stringer)=false",
      "properties": {
        "nanos": {
          "description": "Non-negative fractions of a second at nanosecond resolution. Negative\nsecond values with fractions must still have non-negative nanos values\nthat count forward in time. Must be from 0 to 999,999,999\ninclusive. This field may be limited in precision depending on context.",
          "type": "integer"
        },
        "seconds": {
          "description": "Represents seconds of UTC time since Unix epoch\n1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to\n9999-12-31T23:59:59Z inclusive.",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1API": {
      "properties": {
        "advertiseAddress": {
          "description": "AdvertiseAddress sets the address for the API server to advertise.",
          "type": "string"
        },
        "bindPort": {
          "title": "BindPort sets the secure port for the API Server to bind to",
          "type": "integer"
        }
      },
      "type": "object"
    },
    "v1alpha1AWSSpec": {
      "properties": {
        "iamProfileMaster": {
          "title": "aws:TAG KubernetesCluster => clusterid",
          "type": "string"
        },
        "iamProfileNode": {
          "type": "string"
        },
        "masterIPSuffix": {
          "type": "string"
        },
        "masterSGName": {
          "type": "string"
        },
        "nodeSGName": {
          "type": "string"
        },
        "subnetCidr": {
          "type": "string"
        },
        "vpcCIDR": {
          "type": "string"
        },
        "vpcCIDRBase": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1AWSStatus": {
      "properties": {
        "dhcpOptionsID": {
          "type": "string"
        },
        "igwID": {
          "type": "string"
        },
        "masterSGID": {
          "type": "string"
        },
        "nodeSGID": {
          "type": "string"
        },
        "routeTableID": {
          "type": "string"
        },
        "subnetID": {
          "type": "string"
        },
        "volumeID": {
          "type": "string"
        },
        "vpcID": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1AzureSpec": {
      "properties": {
        "azureStorageAccountName": {
          "type": "string"
        },
        "instanceImageVersion": {
          "type": "string"
        },
        "resourceGroup": {
          "type": "string"
        },
        "rootPassword": {
          "type": "string"
        },
        "routeTableName": {
          "type": "string"
        },
        "securityGroupName": {
          "type": "string"
        },
        "subnetCidr": {
          "type": "string"
        },
        "subnetName": {
          "type": "string"
        },
        "vnetName": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1CloudSpec": {
      "properties": {
        "aws": {
          "$ref": "#/definitions/v1alpha1AWSSpec"
        },
        "azure": {
          "$ref": "#/definitions/v1alpha1AzureSpec"
        },
        "ccmCredentialName": {
          "type": "string"
        },
        "cloudProvider": {
          "type": "string"
        },
        "gce": {
          "$ref": "#/definitions/v1alpha1GoogleSpec"
        },
        "gke": {
          "$ref": "#/definitions/v1alpha1GKESpec"
        },
        "instanceImage": {
          "title": "master needs it for ossec",
          "type": "string"
        },
        "instanceImageProject": {
          "type": "string"
        },
        "linode": {
          "$ref": "#/definitions/v1alpha1LinodeSpec"
        },
        "os": {
          "type": "string"
        },
        "project": {
          "type": "string"
        },
        "region": {
          "type": "string"
        },
        "sshKeyName": {
          "type": "string"
        },
        "zone": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1CloudStatus": {
      "properties": {
        "aws": {
          "$ref": "#/definitions/v1alpha1AWSStatus"
        },
        "sshKeyExternalID": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1Cluster": {
      "properties": {
        "metadata": {
          "$ref": "#/definitions/apismetav1ObjectMeta"
        },
        "spec": {
          "$ref": "#/definitions/v1alpha1ClusterSpec"
        },
        "status": {
          "$ref": "#/definitions/v1alpha1ClusterStatus"
        }
      },
      "type": "object"
    },
    "v1alpha1ClusterSpec": {
      "properties": {
        "api": {
          "$ref": "#/definitions/v1alpha1API"
        },
        "apiServerCertSANs": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "apiServerExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "authorizationModes": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "caCertName": {
          "type": "string"
        },
        "cloud": {
          "$ref": "#/definitions/v1alpha1CloudSpec"
        },
        "controllerManagerExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "credentialName": {
          "type": "string"
        },
        "frontProxyCACertName": {
          "type": "string"
        },
        "kubeletExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "kubernetesVersion": {
          "type": "string"
        },
        "locked": {
          "type": "boolean"
        },
        "networking": {
          "$ref": "#/definitions/v1alpha1Networking"
        },
        "schedulerExtraArgs": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        }
      },
      "type": "object"
    },
    "v1alpha1ClusterStatus": {
      "properties": {
        "apiServer": {
          "items": {
            "$ref": "#/definitions/v1NodeAddress"
          },
          "type": "array"
        },
        "cloud": {
          "$ref": "#/definitions/v1alpha1CloudStatus"
        },
        "phase": {
          "type": "string"
        },
        "reason": {
          "type": "string"
        },
        "reservedIP": {
          "items": {
            "$ref": "#/definitions/v1alpha1ReservedIP"
          },
          "type": "array"
        }
      },
      "type": "object"
    },
    "v1alpha1GKESpec": {
      "properties": {
        "networkName": {
          "type": "string"
        },
        "password": {
          "type": "string"
        },
        "userName": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1GoogleSpec": {
      "properties": {
        "networkName": {
          "type": "string"
        },
        "nodeScopes": {
          "items": {
            "type": "string"
          },
          "title": "gce\nNODE_SCOPES=\"${NODE_SCOPES:-compute-rw,monitoring,logging-write,storage-ro}\"",
          "type": "array"
        },
        "nodeTags": {
          "items": {
            "type": "string"
          },
          "type": "array"
        }
      },
      "type": "object"
    },
    "v1alpha1LinodeSpec": {
      "properties": {
        "kernelId": {
          "type": "integer"
        },
        "rootPassword": {
          "title": "Linode",
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1Networking": {
      "properties": {
        "dnsDomain": {
          "type": "string"
        },
        "dnsServerIP": {
          "title": "NEW\nReplacing API_SERVERS https://github.com/kubernetes/kubernetes/blob/62898319dff291843e53b7839c6cde14ee5d2aa4/cluster/aws/util.sh#L1004",
          "type": "string"
        },
        "masterSubnet": {
          "type": "string"
        },
        "networkProvider": {
          "type": "string"
        },
        "nonMasqueradeCIDR": {
          "type": "string"
        },
        "podSubnet": {
          "title": "kubenet, flannel, calico, opencontrail",
          "type": "string"
        },
        "serviceSubnet": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "v1alpha1ReservedIP": {
      "properties": {
        "id": {
          "type": "string"
        },
        "ip": {
          "type": "string"
        },
        "name": {
          "type": "string"
        }
      },
      "type": "object"
    }
  },
  "properties": {
    "cluster": {
      "$ref": "#/definitions/v1alpha1Cluster"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
	clusterMetadataRequestSchema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "properties": {
    "uid": {
      "type": "string"
    }
  },
  "type": "object"
}`))
	if err != nil {
		glog.Fatal(err)
	}
}

func (m *SSHConfigGetRequest) Valid() (*gojsonschema.Result, error) {
	return sSHConfigGetRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *SSHConfigGetRequest) IsRequest() {}

func (m *ClusterApplyRequest) Valid() (*gojsonschema.Result, error) {
	return clusterApplyRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ClusterApplyRequest) IsRequest() {}

func (m *CredentialCreateRequest) Valid() (*gojsonschema.Result, error) {
	return credentialCreateRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *CredentialCreateRequest) IsRequest() {}

func (m *NodeGroupDeleteRequest) Valid() (*gojsonschema.Result, error) {
	return nodeGroupDeleteRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *NodeGroupDeleteRequest) IsRequest() {}

func (m *ClusterUpdateRequest) Valid() (*gojsonschema.Result, error) {
	return clusterUpdateRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ClusterUpdateRequest) IsRequest() {}

func (m *NodeGroupListRequest) Valid() (*gojsonschema.Result, error) {
	return nodeGroupListRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *NodeGroupListRequest) IsRequest() {}

func (m *ClusterDeleteRequest) Valid() (*gojsonschema.Result, error) {
	return clusterDeleteRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ClusterDeleteRequest) IsRequest() {}

func (m *NodeGroupDescribeRequest) Valid() (*gojsonschema.Result, error) {
	return nodeGroupDescribeRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *NodeGroupDescribeRequest) IsRequest() {}

func (m *CredentialDescribeRequest) Valid() (*gojsonschema.Result, error) {
	return credentialDescribeRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *CredentialDescribeRequest) IsRequest() {}

func (m *CredentialDeleteRequest) Valid() (*gojsonschema.Result, error) {
	return credentialDeleteRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *CredentialDeleteRequest) IsRequest() {}

func (m *ClusterDescribeRequest) Valid() (*gojsonschema.Result, error) {
	return clusterDescribeRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ClusterDescribeRequest) IsRequest() {}

func (m *ClusterListRequest) Valid() (*gojsonschema.Result, error) {
	return clusterListRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ClusterListRequest) IsRequest() {}

func (m *ClusterClientConfigRequest) Valid() (*gojsonschema.Result, error) {
	return clusterClientConfigRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ClusterClientConfigRequest) IsRequest() {}

func (m *NodeGroupUpdateRequest) Valid() (*gojsonschema.Result, error) {
	return nodeGroupUpdateRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *NodeGroupUpdateRequest) IsRequest() {}

func (m *NodeGroupCreateRequest) Valid() (*gojsonschema.Result, error) {
	return nodeGroupCreateRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *NodeGroupCreateRequest) IsRequest() {}

func (m *CredentialUpdateRequest) Valid() (*gojsonschema.Result, error) {
	return credentialUpdateRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *CredentialUpdateRequest) IsRequest() {}

func (m *ClusterCreateRequest) Valid() (*gojsonschema.Result, error) {
	return clusterCreateRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ClusterCreateRequest) IsRequest() {}

func (m *ClusterMetadataRequest) Valid() (*gojsonschema.Result, error) {
	return clusterMetadataRequestSchema.Validate(gojsonschema.NewGoLoader(m))
}
func (m *ClusterMetadataRequest) IsRequest() {}


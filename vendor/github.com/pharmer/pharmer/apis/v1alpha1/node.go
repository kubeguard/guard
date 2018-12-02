package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeNodeGroup = "ng"
	ResourceKindNodeGroup = "NodeGroup"
	ResourceNameNodeGroup = "nodegroup"
	ResourceTypeNodeGroup = "nodegroups"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              NodeGroupSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            NodeGroupStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type NodeGroupSpec struct {
	Nodes int64 `json:"nodes" protobuf:"varint,1,opt,name=nodes"`
	// Template describes the nodes that will be created.
	Template NodeTemplateSpec `json:"template" protobuf:"bytes,2,opt,name=template"`
}

// NodeGroupStatus is the most recently observed status of the NodeGroup.
type NodeGroupStatus struct {
	// Nodes is the most recently oberved number of nodes.
	Nodes int64 `json:"nodes" protobuf:"varint,1,opt,name=nodes"`
	// ObservedGeneration reflects the generation of the most recently observed node group.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`
}

func (ng NodeGroup) IsMaster() bool {
	_, found := ng.Labels[RoleMasterKey]
	return found
}

func (ng NodeGroup) Role() string {
	if ng.IsMaster() {
		return RoleMaster
	}
	return RoleNode
}

// PodTemplateSpec describes the data a pod should have when created from a template
type NodeTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	// metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec NodeSpec `json:"spec,omitempty" protobuf:"bytes,1,opt,name=spec"`
}

type IPType string

const (
	IPTypeEphemeral IPType = "Ephemeral"
	IPTypeReserved  IPType = "Reserved"
)

type NodeType string

const (
	NodeTypeRegular NodeType = "regular"
	NodeTypeSpot    NodeType = "spot"
)

type NodeSpec struct {
	SKU              string            `json:"sku,omitempty" protobuf:"bytes,1,opt,name=sku"`
	DiskType         string            `json:"nodeDiskType,omitempty" protobuf:"bytes,2,opt,name=nodeDiskType"`
	DiskSize         int64             `json:"nodeDiskSize,omitempty" protobuf:"varint,3,opt,name=nodeDiskSize"`
	ExternalIPType   IPType            `json:"externalIPType,omitempty" protobuf:"bytes,4,opt,name=externalIPType,casttype=IPType"`
	KubeletExtraArgs map[string]string `json:"kubeletExtraArgs,omitempty" protobuf:"bytes,5,rep,name=kubeletExtraArgs"`
	Type             NodeType          `json:"type,omitempty" protobuf:"varint,6,opt,name=type"`
	SpotPriceMax     float64           `json:"spotPriceMax,omitempty" protobuf:"fixed64,7,opt,name=spotPriceMax"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeInfo struct {
	metav1.TypeMeta `json:",inline"`
	Name            string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	ExternalID      string `json:"externalID,omitempty" protobuf:"bytes,2,opt,name=externalID"`
	PublicIP        string `json:"publicIP,omitempty" protobuf:"bytes,3,opt,name=publicIP"`
	PrivateIP       string `json:"privateIP,omitempty" protobuf:"bytes,4,opt,name=privateIP"`
	DiskId          string `json:"diskID,omitempty" protobuf:"bytes,5,opt,name=diskID"`
}

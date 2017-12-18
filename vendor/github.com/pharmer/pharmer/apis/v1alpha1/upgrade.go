package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//https://github.com/kubernetes/kubernetes/blob/aa1dc9db3532dfbf09e45c8e3786a648cd217417/cmd/kubeadm/app/phases/upgrade/compute.go#L28
type Upgrade struct {
	metav1.TypeMeta `json:",inline,omitempty,omitempty"`

	Description string       `json:"description" protobuf:"bytes,1,opt,name=description"`
	Before      ClusterState `json:"before" protobuf:"bytes,2,opt,name=before"`
	After       ClusterState `json:"after" protobuf:"bytes,3,opt,name=after"`
}

// ClusterState describes the state of certain versions for a cluster
type ClusterState struct {
	// KubeVersion describes the version of the Kubernetes API Server, Controller Manager, Scheduler and Proxy.
	KubeVersion string `json:"kubeVersion" protobuf:"bytes,1,opt,name=kubeVersion"`
	// DNSVersion describes the version of the kube-dns images used and manifest version
	DNSVersion string `json:"dnsVersion" protobuf:"bytes,2,opt,name=dnsVersion"`
	// MasterKubeadmVersion describes the version of the kubeadm CLI
	KubeadmVersion string `json:"kubeadmVersion" protobuf:"bytes,3,opt,name=kubeadmVersion"`
	// KubeletVersions is a map with a version number linked to the amount of kubelets running that version in the cluster
	KubeletVersions map[string]uint32 `json:"kubeletVersions" protobuf:"bytes,4,rep,name=kubeletVersions"`
}

// CanUpgradeKubelets returns whether an upgrade of any kubelet in the cluster is possible
func (u *Upgrade) CanUpgradeKubelets() bool {
	// If there are multiple different versions now, an upgrade is possible (even if only for a subset of the nodes)
	if len(u.Before.KubeletVersions) > 1 {
		return true
	}
	// Don't report something available for upgrade if we don't know the current state
	if len(u.Before.KubeletVersions) == 0 {
		return false
	}

	// if the same version number existed both before and after, we don't have to upgrade it
	_, sameVersionFound := u.Before.KubeletVersions[u.After.KubeVersion]
	return !sameVersionFound
}

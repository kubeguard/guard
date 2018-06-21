package framework

import (
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) DeleteClusterRole(name string) error {
	return f.KubeClient.RbacV1().ClusterRoles().Delete(name, deleteInBackground())
}

func (f *Framework) GetClusterRole(name string) (*rbac.ClusterRole, error) {
	return f.KubeClient.RbacV1().ClusterRoles().Get(name, metav1.GetOptions{})
}

func (f *Framework) DeleteClusterRoleBinding(name string) error {
	return f.KubeClient.RbacV1().ClusterRoleBindings().Delete(name, deleteInBackground())
}

func (f *Framework) GetClusterRoleBinding(name string) (*rbac.ClusterRoleBinding, error) {
	return f.KubeClient.RbacV1().ClusterRoleBindings().Get(name, metav1.GetOptions{})
}

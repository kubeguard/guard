/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

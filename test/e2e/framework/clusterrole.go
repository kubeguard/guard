package framework

func (f *Framework) DeleteClusterRole(name string) error {
	return f.KubeClient.RbacV1().ClusterRoles().Delete(name, deleteInBackground())
}

func (f *Framework) DeleteClusterRoleBinding(name string) error {
	return f.KubeClient.RbacV1().ClusterRoleBindings().Delete(name, deleteInBackground())
}

package framework

func (f *Framework) DeleteDeployment(name, namespace string) error {
	return f.KubeClient.AppsV1beta1().Deployments(namespace).Delete(name, deleteInBackground())
}

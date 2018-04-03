package framework

func (f *Framework) DeleteSecret(name, namespace string) error {
	return f.KubeClient.CoreV1().Secrets(namespace).Delete(name, deleteInBackground())
}

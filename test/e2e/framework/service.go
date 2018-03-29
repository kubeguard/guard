package framework

func (f *Framework) DeleteService(name, namespace string) error {
	return f.KubeClient.CoreV1().Services(namespace).Delete(name, deleteInForeground())
}

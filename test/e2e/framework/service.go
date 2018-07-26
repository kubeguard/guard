package framework

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) DeleteService(name, namespace string) error {
	return f.KubeClient.CoreV1().Services(namespace).Delete(name, deleteInForeground())
}

func (f *Framework) GetService(name, namespace string) (*corev1.Service, error) {
	return f.KubeClient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
}

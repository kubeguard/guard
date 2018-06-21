package framework

import (
	apps "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) DeleteDeployment(name, namespace string) error {
	return f.KubeClient.AppsV1beta1().Deployments(namespace).Delete(name, deleteInBackground())
}

func (f *Framework) GetDeployment(name, namespace string) (*apps.Deployment, error) {
	return f.KubeClient.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

package framework

import (
	"os/exec"
	"time"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	updateRetryInterval = 10 * 1000 * 1000 * time.Nanosecond
	maxAttempts         = 5
)

func deleteInBackground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationBackground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func deleteInForeground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationForeground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func CleanupMinikubeHostPath() error {
	cmd := "minikube"
	args := []string{"ssh", "sudo rm -rf /data/guard-test"}
	return exec.Command(cmd, args...).Run()
}

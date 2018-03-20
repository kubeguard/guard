package installer

import (
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func newServiceAccount(namespace string) runtime.Object {
	return &core.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func newClusterRole(namespace string) runtime.Object {
	return &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{core.GroupName},
				Resources: []string{"nodes"},
				Verbs:     []string{"list"},
			},
		},
	}
}

func newClusterRoleBinding(namespace string) runtime.Object {
	return &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     "guard",
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Name:      "guard",
				Namespace: namespace,
			},
		},
	}
}

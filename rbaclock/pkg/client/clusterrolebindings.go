package client

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListClusterRoleBindings(namespace string) []v1.ClusterRoleBinding {
	result := []v1.ClusterRoleBinding{}
	if namespace == "" {
		clusterrolebindings, _ := Client().RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
		for _, crb := range clusterrolebindings.Items {
			if len(crb.Subjects) == 0 {
				continue
			}
			result = append(result, crb)
		}
	} else {
		clusterrolebindings, _ := Client().RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
		for _, crb := range clusterrolebindings.Items {
			if len(crb.Subjects) == 0 {
				continue
			}
			for _, subject := range crb.Subjects {
				if subject.Kind == "ServiceAccount" && subject.Namespace == namespace {
					result = append(result, crb)
				}
			}
		}
	}

	return result
}

func ClusterRoleBinding(name string) *v1.ClusterRoleBinding {
	clusterrolebinding, _ := Client().RbacV1().ClusterRoleBindings().Get(context.TODO(), name, metav1.GetOptions{})
	return clusterrolebinding
}

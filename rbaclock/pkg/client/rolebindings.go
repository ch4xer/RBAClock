package client

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListRoleBindings(namespace string) []v1.RoleBinding {
	result := []v1.RoleBinding{}
	if namespace == "" {
		namespaces := ListNamespaces()
		for _, ns := range namespaces {
			rolebindings, _ := Client().RbacV1().RoleBindings(ns.Name).List(context.TODO(), metav1.ListOptions{})
			for _, rb := range rolebindings.Items {
				if len(rb.Subjects) == 0 {
					continue
				}
				result = append(result, rb)
			}
		}
	} else {
		rolebindings, _ := Client().RbacV1().RoleBindings(namespace).List(context.TODO(), metav1.ListOptions{})
		for _, rb := range rolebindings.Items {
			if len(rb.Subjects) == 0 {
				continue
			}
			result = append(result, rb)
		}
	}

	return result
}

func RoleBinding(namespace, name string) *v1.RoleBinding {
	rolebinding, _ := Client().RbacV1().RoleBindings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return rolebinding
}

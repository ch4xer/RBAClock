package client

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// kubernetes.io/service-account-token
func ListSecrets(namespace, secrestType string) []v1.Secret {
	result := []v1.Secret{}
	if namespace == "" {
		namespaces := ListNamespaces()
		for _, ns := range namespaces {
			secrets, _ := Client().CoreV1().Secrets(ns.Namespace).List(context.TODO(), metav1.ListOptions{
				FieldSelector: fmt.Sprintf("type=%s", secrestType),
			})
			result = append(result, secrets.Items...)
		}
	} else {
		secrets, _ := Client().CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("type=%s", secrestType),
		})
		result = append(result, secrets.Items...)
	}

	return result
}

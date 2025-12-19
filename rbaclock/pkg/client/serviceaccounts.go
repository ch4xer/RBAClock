package client

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListServiceAccounts(namespace string) []v1.ServiceAccount {
	sas, _ := Client().CoreV1().ServiceAccounts(namespace).List(context.TODO(), metav1.ListOptions{})
	return sas.Items
}

func ServiceAccount(namespace, name string) *v1.ServiceAccount {
	sa, _ := Client().CoreV1().ServiceAccounts(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return sa
}

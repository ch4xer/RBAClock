package client

import (
	"context"

	"rbaclock/conf"

	log "github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"slices"
)

func Namespace(name string) *v1.Namespace {
	namespace, err := Client().CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.V(2).Infof("Failed to get namespace %s: %v", name, err)
	}
	return namespace
}

func ListNamespaces() []v1.Namespace {
	namespaces, err := Client().CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.V(2).Infof("Failed to list namespace: %v", err)
	}
	result := []v1.Namespace{}
	for _, namespace := range namespaces.Items {
		if shouldIgnoreNamespace(namespace.Name) {
			continue
		}
		result = append(result, namespace)
	}

	return result
}

func ListNamespacesExceptKubeSystem() []v1.Namespace {
	namespaces := ListNamespaces()
	result := []v1.Namespace{}
	for _, namespace := range namespaces {
		if namespace.Name != "kube-system" {
			result = append(result, namespace)
		}
	}
	return result
}

func shouldIgnoreNamespace(namespace string) bool {
	return slices.Contains(conf.IgnoreNamespaces, namespace)
}

func RemoveNamespaceToleration(namespace, key string) {
	pods := Pods(namespace)
	for _, pod := range pods {
		if len(pod.OwnerReferences) == 0 {
			removePodToleration(pod, key)
		} else {
			removeOwnerToleration(Owner(pod), key)
		}

	}
}

func RemoveNamespaceAffinity(namespace, key string) {
	pods := Pods(namespace)
	for _, pod := range pods {
		if len(pod.OwnerReferences) == 0 {
			removePodAffinity(pod, key)
		} else {
			removeOwnerAffinity(Owner(pod), key)
		}
	}
}

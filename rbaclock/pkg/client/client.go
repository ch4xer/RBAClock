package client

import (

	"rbaclock/conf"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Client() *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", conf.Kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset
}

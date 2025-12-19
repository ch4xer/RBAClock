package scan

import (
	cli "rbaclock/pkg/client"
	priv "rbaclock/pkg/privilege"
	"slices"

	v1 "k8s.io/api/core/v1"
)

type Priv = priv.Privilege

func TakeNamespaceRolePrivs(namespace string) []Priv {
	privs := []Priv{}
	roles := cli.ListRoles(namespace)
	for _, role := range roles {
		privs = append(privs, priv.PrivEntry(role.Rules, namespace)...)
	}
	return privs
}

// means that all namespaced sa are accessible, except kube-system
func TakeOverNodeSaPrivs() map[string][]Priv {
	saPrivs := map[string][]Priv{}
	namespaces := cli.ListNamespacesExceptKubeSystem()
	for _, ns := range namespaces {
		nsSaPrivs := TakeNamespaceSaPrivs(ns.Name)
		for k, v := range nsSaPrivs {
			saPrivs[k] = v
		}
	}
	return saPrivs
}

func TakeNodeSaPrivs(node *v1.Node) map[string][]Priv {
	saPrivs := map[string][]Priv{}
	pods := cli.PodsOnNode(node)
	for _, pod := range pods {
		index := pod.Namespace + ":" + pod.Spec.ServiceAccountName
		if _, ok := saPrivs[index]; !ok {
			saPrivs[index] = []Priv{}
		}
		roles := cli.RolesBoundToServiceAccount(pod.Namespace, pod.Spec.ServiceAccountName)
		clusterroles := cli.ClusterRolesBoundToServiceAccount(pod.Namespace, pod.Spec.ServiceAccountName)
		for _, role := range roles {
			saPrivs[index] = priv.MergePrivs(
				saPrivs[index],
				priv.PrivEntry(role.Rules, pod.Namespace),
			)
		}
		for _, clusterrole := range clusterroles {
			saPrivs[index] = priv.MergePrivs(
				saPrivs[index],
				priv.PrivEntry(clusterrole.Rules, ""),
			)
		}
	}
	return saPrivs
}

func TakeNamespaceSaPrivs(namespace string) map[string][]Priv {
	// namespace:sa -> privs
	saPrivs := map[string][]Priv{}

	serviceAccounts := cli.ListServiceAccounts(namespace)
	for _, sa := range serviceAccounts {
		index := namespace + ":" + sa.Name
		saPrivs[index] = []Priv{}
		roles := cli.RolesBoundToServiceAccount(namespace, sa.Name)
		clusterroles := cli.ClusterRolesBoundToServiceAccount(namespace, sa.Name)
		for _, role := range roles {
			saPrivs[index] = priv.MergePrivs(
				saPrivs[index],
				priv.PrivEntry(role.Rules, namespace),
			)
		}
		for _, clusterrole := range clusterroles {
			saPrivs[index] = priv.MergePrivs(
				saPrivs[index],
				priv.PrivEntry(clusterrole.Rules, ""),
			)
		}
	}
	return saPrivs
}

func TakeNamespaceSaTokenPrivs(namespace string) map[string][]Priv {
	privsMap := map[string][]Priv{}
	secrets := cli.ListSecrets(namespace, "kubernetes.io/service-account-token")

	roleBindings := cli.ListRoleBindings(namespace)
	clusterroleBindings := cli.ListClusterRoleBindings(namespace)

	for _, secret := range secrets {
		sa := secret.Annotations["kubernetes.io/service-account.name"]

		for _, rb := range roleBindings {
			if rb.Subjects[0].Name == sa && rb.Subjects[0].Kind == "ServiceAccount" {
				index := namespace + ":" + sa
				if _, ok := privsMap[index]; !ok {
					privsMap[index] = []Priv{}
				}
				roles := cli.RolesBoundToServiceAccount(namespace, sa)
				for _, role := range roles {
					privsMap[index] = append(privsMap[index], priv.PrivEntry(role.Rules, namespace)...)
				}

			}
		}

		for _, crb := range clusterroleBindings {
			if crb.Subjects[0].Name == sa && crb.Subjects[0].Kind == "ServiceAccount" {
				index := namespace + ":" + sa
				if _, ok := privsMap[index]; !ok {
					privsMap[index] = []Priv{}
				}
				clusterrole := cli.ClusterRole(crb.RoleRef.Name)
				privsMap[index] = append(privsMap[index], priv.PrivEntry(clusterrole.Rules, "")...)
			}
		}
	}
	return privsMap
}

func TakeClusterSaTokenPrivs() map[string][]Priv {
	// namespace:sa -> privs
	privsMap := map[string][]Priv{}

	secrets := cli.ListSecrets("", "kubernetes.io/service-account-token")
	for _, secret := range secrets {
		sa := secret.Annotations["kubernetes.io/service-account.name"]
		index := secret.Namespace + ":" + sa
		if _, ok := privsMap[index]; !ok {
			privsMap[index] = []Priv{}
		}
		roles := cli.RolesBoundToServiceAccount(secret.Namespace, sa)
		clusterroles := cli.ClusterRolesBoundToServiceAccount(secret.Namespace, sa)
		for _, role := range roles {
			privsMap[index] = append(privsMap[index], priv.PrivEntry(role.Rules, secret.Namespace)...)
		}
		for _, clusterrole := range clusterroles {
			privsMap[index] = append(privsMap[index], priv.PrivEntry(clusterrole.Rules, "")...)
		}
	}

	return privsMap
}

func TakeNamespacePodPrivs(namespace string) map[string][]Priv {
	privs := map[string][]Priv{}
	pods := cli.Pods(namespace)
	for _, pod := range pods {
		index := pod.Namespace + ":" + pod.Spec.ServiceAccountName
		if _, ok := privs[index]; !ok {
			privs[index] = []Priv{}
		}
		roles := cli.RolesBoundToServiceAccount(namespace, pod.Spec.ServiceAccountName)
		clusterroles := cli.ClusterRolesBoundToServiceAccount(namespace, pod.Spec.ServiceAccountName)
		for _, role := range roles {
			privs[index] = append(privs[index], priv.PrivEntry(role.Rules, pod.Namespace)...)
		}
		for _, clusterrole := range clusterroles {
			privs[index] = append(privs[index], priv.PrivEntry(clusterrole.Rules, pod.Namespace)...)
		}
	}
	return privs
}

func inArray(item string, arr []string) bool {
	return slices.Contains(arr, item)
}

func Filter(privsMap map[string][]Priv, exclude []string) map[string][]Priv {
	filtered := map[string][]Priv{}
	for k, v := range privsMap {
		if inArray(k, exclude) {
			continue
		}
		filtered[k] = v
	}
	return filtered
}

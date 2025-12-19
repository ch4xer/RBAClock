package client

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListClusterRoles() []rbacv1.ClusterRole {
	clusterroles, _ := Client().RbacV1().ClusterRoles().List(context.TODO(), metav1.ListOptions{})
	return clusterroles.Items
}

func ClusterRole(name string) *rbacv1.ClusterRole {
	clusterrole, _ := Client().RbacV1().ClusterRoles().Get(context.TODO(), name, metav1.GetOptions{})
	return clusterrole
}

func ClusterRolesBoundToServiceAccount(namespace, serviceAccountName string) []rbacv1.ClusterRole {
	var boundClusterRoles []rbacv1.ClusterRole
	clusterRoleBindings := ListClusterRoleBindings(namespace)
	for _, crb := range clusterRoleBindings {
		for _, subject := range crb.Subjects {
			if subject.Kind == "ServiceAccount" && subject.Name == serviceAccountName {

				clusterRole, _ := Client().RbacV1().ClusterRoles().Get(context.TODO(), crb.RoleRef.Name, metav1.GetOptions{})

				boundClusterRoles = append(boundClusterRoles, *clusterRole)
			}
		}
	}
	return boundClusterRoles
}

func FindClusterRolesBoundToUser(groups []string, user string) []*rbacv1.ClusterRole {
	result := []*rbacv1.ClusterRole{}
	clusterRoleBindings := ListClusterRoleBindings("")
	for _, crb := range clusterRoleBindings {
		for _, subject := range crb.Subjects {
			if subject.Kind == "User" && subject.Name == user {
				clusterrole, _ := Client().RbacV1().ClusterRoles().Get(context.TODO(), crb.RoleRef.Name, metav1.GetOptions{})
				result = append(result, clusterrole)
			} else if subject.Kind == "Group" {
				for _, group := range groups {
					if subject.Name == group {
						clusterrole, _ := Client().RbacV1().ClusterRoles().Get(context.TODO(), crb.RoleRef.Name, metav1.GetOptions{})
						result = append(result, clusterrole)
					}
				}
			}
		}
	}
	return result
}


func FindUserClusterRole(clusterRoleBindings []rbacv1.ClusterRoleBinding, user string, groups []string) []*rbacv1.ClusterRole {
	result := []*rbacv1.ClusterRole{}
	for _, crb := range clusterRoleBindings {
		for _, subject := range crb.Subjects {
			if subject.Kind == "User" && subject.Name == user {
				clusterrole, _ := Client().RbacV1().ClusterRoles().Get(context.TODO(), crb.RoleRef.Name, metav1.GetOptions{})
				result = append(result, clusterrole)
			}
			if subject.Kind == "Group" {
				for _, group := range groups {
					if subject.Name == group {
						clusterrole, _ := Client().RbacV1().ClusterRoles().Get(context.TODO(), crb.RoleRef.Name, metav1.GetOptions{})
						result = append(result, clusterrole)
					}
				}
			}
		}
	}
	return result
}

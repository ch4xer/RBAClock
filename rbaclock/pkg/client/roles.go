package client

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListRoles(namespace string) []rbacv1.Role {
	roles, _ := Client().RbacV1().Roles(namespace).List(context.TODO(), metav1.ListOptions{})
	return roles.Items
}

func Role(namespace, name string) *rbacv1.Role {
	role, _ := Client().RbacV1().Roles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return role
}

func RolesBoundToServiceAccount(namespace, serviceAccountName string) []rbacv1.Role {
	var boundRoles []rbacv1.Role
	roleBindings := ListRoleBindings(namespace)

	for _, rb := range roleBindings {
		for _, subject := range rb.Subjects {
			if subject.Kind == "ServiceAccount" && subject.Name == serviceAccountName {

				role, _ := Client().RbacV1().Roles(namespace).Get(context.TODO(), rb.RoleRef.Name, metav1.GetOptions{})

				boundRoles = append(boundRoles, *role)
			}
		}
	}
	return boundRoles
}

func FindRolesBoundToUser(groups []string, user string ) []*rbacv1.Role {
	result := []*rbacv1.Role{}
	roleBindings := ListRoleBindings("")
	for _, rb := range roleBindings {
		for _, subject := range rb.Subjects {
			if subject.Kind == "User" && subject.Name == user {
				role, _ := Client().RbacV1().Roles(rb.Namespace).Get(context.TODO(), rb.RoleRef.Name, metav1.GetOptions{})
				result = append(result, role)
			}
			if subject.Kind == "Group" {
				for _, group := range groups {
					if subject.Name == group {
						role, _ := Client().RbacV1().Roles(rb.Namespace).Get(context.TODO(), rb.RoleRef.Name, metav1.GetOptions{})
						result = append(result, role)
					}
				}
			}
		}
	}
	return result
}

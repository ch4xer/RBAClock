# RBAClock: Contain RBAC Permissions through Secure Scheduling

RBAClock is a prototype framework to prevent RBAC escalation attacks in Kubernetes by enforcing scheduling constraints based on Role-Based Access Control (RBAC) permissions. 

The basic idea is to group pods with RBAC permissions that have similar impacts on the cluster and schedule them on the same set of nodes. This way, even if an attacker compromises a pod and escalates RBAC privileges with container escalation, the attacker's benefit is limited to the similar RBAC permissions on that node group.

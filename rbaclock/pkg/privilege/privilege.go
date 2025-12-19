package privilege

import (
	"rbaclock/pkg/db"
	"strings"
)

// namespace:resource:verb
type Privilege string

type Knowledge struct {
	Conditions []Privilege
	Effect     func() []Privilege
}

type Risk struct {
	Severity int
	Name     string
	Desc     string
}

var (
	NoRisk Risk = Risk{0, "NoRisk", "No risk"}

	TakeNamespaceSaTokens Risk = Risk{1, "TakeSaTokens", "Potential to check service account tokens in specific namespaces"}
	// // TakeClusterPod == TakeOverCluster
	// TakeNamespacePods Risk = Risk{10, "TakePods", "Potential to take over pods in specific namespaces"}
	// TakeClusterRole == TakeOverCluster
	TakeNamespaceRoles Risk = Risk{1, "TakeRoles", "Potential to take over roles in specific namespaces"}
	// TakeClusterSa == TakeOverCluster
	// sa may bound to cluster-level roles, so more dangerous than namespace-level roles
	TakeNamespaceSa Risk = Risk{1, "TakeSa", "Potential to take over service accounts in specific namespaces"}
	// TakeClusterSaToken Risk = Risk{1, "TakeClusterSaToken", "Potential to check service account tokens in all namespaces"}
	// PrivEscalation     Risk = Risk{1, "PrivEscalation", "Potential to perform privilege escalation"}
	LeakInfo                  Risk = Risk{10, "LeakInfo", "Potential to leak sensitive information"}
	DestroyNamespaceContainer Risk = Risk{100, "DestroyNamespaceResource", "Potential to destroy resources in a namespace"}
	DestroyClusterContainer   Risk = Risk{1000, "DestroyClusterContainer", "Potential to destroy cluster container"}
	DestroyCurrentNode        Risk = Risk{1000, "DestroyCurrentNode", "Potential to destroy current node"}
	TakeOverPod               Risk = Risk{10000, "TakeOverPod", "Potential to take over one pod in cluster"}
	TakeOverNode              Risk = Risk{100000, "TakeOverNode", "Potential to take over one node in cluster"}
	TakeOverCluster           Risk = Risk{1000000, "TakeOverCluster", "Potential to take over the cluster"}
)

var KnowledgeBase map[Risk][][]Privilege = knowledge()

func knowledge() map[Risk][][]Privilege {
	knowledge := make(map[Risk][][]Privilege)
	knowledge[TakeOverCluster] = [][]Privilege{
		{"*:mutatingwebhookconfigurations:create,patch,update"},
		{
			"*:roles:escalate",
			"*:roles:patch",
		},
		{
			"*:clusterroles:escalate",
			"*:clusterroles:patch",
		},
		{
			"*:rolebindings:create,patch,update",
			"*:clusterroles:bind",
		},
		{
			"*:clusterrolebindings:create,patch,update",
			"*:clusterroles:bind",
		},
		{"cluster:serviceaccounts:impersonate"},
		{"cluster:serviceaccounts/token:create"},
		{"cluster:secrets:create", "cluster:secrets:list,get"},
		{"cluster:secrets:list,get"},
		{"cluster:pods/exec:create"},
		{"cluster:users:impersonate", "cluster:groups:impersonate"},
	}

	knowledge[TakeOverNode] = [][]Privilege{
		{"*:pods:create,patch,update"},
		{"*:statefulsets:create,patch,update"},
		{"*:replicasets:create,patch,update"},
		{"*:jobs:create,patch,update"},
		{"*:deployments:create,patch,update"},
		{"*:daemonsets:create,patch,update"},
		{"*:cronjobs:create,patch,update"},
	}

	knowledge[TakeOverPod] = [][]Privilege{
		{"namespace:pods/exec:create"},
	}

	knowledge[DestroyNamespaceContainer] = [][]Privilege{
		{"namespace:pods:delete"},
		{"namespace:statefulsets:delete"},
		{"namespace:replicasets:delete"},
		{"namespace:jobs:delete"},
		{"namespace:deployments:delete"},
		{"namespace:daemonsets:delete"},
		{"namespace:cronjobs:delete"},
		{"namespace:ingresses:delete"},
		{"namespace:pods/eviction:create"},
		{"namespace:services:update,patch,delete"},
		{"namespace:networkpolicies:create,update,patch,delete"},
	}

	knowledge[DestroyClusterContainer] = [][]Privilege{
		{"cluster:pods:delete"},
		{"cluster:statefulsets:delete"},
		{"cluster:replicasets:delete"},
		{"cluster:jobs:delete"},
		{"cluster:deployments:delete"},
		{"cluster:daemonsets:delete"},
		{"cluster:cronjobs:delete"},
		{"cluster:ingresses:delete"},
		{"cluster:pods/eviction:create"},
		{"cluster:services:update,patch,delete"},
		{"cluster:networkpolicies:create,update,patch,delete"},
		{"cluster:validatingwebhookconfigurations:create,update,patch"},
	}

	knowledge[DestroyCurrentNode] = [][]Privilege{
		{"*:nodes:update,patch,delete"},
	}

	knowledge[LeakInfo] = [][]Privilege{
		{"namespace:secrets:get,list"},
		{"*:services:update,patch,create"},
		{"*:networkpolicies:create,update,patch,delete"},
		{"*:configmaps:get,list"},
	}

	// NOTE: this should be broken down into multiple cases
	// knowledge[PrivEscalation] = [][]Privilege{
	// 	{"namespace:secrets:get,list"},
	// 	{"namespace:serviceaccounts:impersonate"},
	// 	{"namespace:serviceaccounts/token:create"},
	// 	{
	// 		"namespace:rolebindings:create,patch,update",
	// 		"namespace:roles:bind",
	// 	},
	// 	{
	// 		"namespace:secrets:create",
	// 		"namespace:secrets:list,get",
	// 	},
	// }

	// Sa can't bind to roles in another namespace
	knowledge[TakeNamespaceRoles] = [][]Privilege{
		{
			"namespace:rolebindings:create,patch,update",
			"namespace:roles:bind",
		},
	}

	knowledge[TakeNamespaceSa] = [][]Privilege{
		{"namespace:serviceaccounts:impersonate"},
		{"namespace:serviceaccounts/token:create"},
		{"namespace:secrets:create", "namespace:secrets:list,get"},
	}

	knowledge[TakeNamespaceSaTokens] = [][]Privilege{
		{"namespace:secrets:list,get"},
	}

	return knowledge
}

func PrivRisks(privs []Privilege) []Risk {
	risks := []Risk{}
	nsPrivs := map[string][]Privilege{}
	for _, priv := range privs {
		parts := strings.Split(string(priv), ":")
		namespace := parts[0]
		resource := parts[1]
		// in case of custom resource
		builtins := db.QueryCRMap(resource)
		verb := parts[2]
		if _, ok := nsPrivs[namespace]; !ok {
			nsPrivs[namespace] = []Privilege{}
		}
		var scope string
		if namespace == "cluster" {
			scope = "cluster"
		} else {
			scope = "namespace"
		}
		if len(builtins) == 0 {
			nsPrivs[namespace] = append(nsPrivs[namespace], Privilege(scope+":"+resource+":"+verb))
		} else {
			for _, builtin := range builtins {
				nsPrivs[namespace] = append(nsPrivs[namespace], Privilege(scope+":"+builtin+":"+verb))
			}
		}
	}

	sortedRisks := []Risk{TakeOverCluster, TakeNamespaceRoles, TakeNamespaceSa, TakeNamespaceSaTokens, TakeOverNode, TakeOverPod, DestroyCurrentNode, DestroyClusterContainer, DestroyNamespaceContainer, LeakInfo}
	for _, privs := range nsPrivs {
		for _, risk := range sortedRisks {
			for _, privPattern := range KnowledgeBase[risk] {
				if isSubsetPrivs(privPattern, privs) {
					risks = MergeRisk(risks, risk)
				}
			}
		}
	}
	result := Process(risks)
	return result
}

func Process(risks []Risk) []Risk {
	// bubble sort
	for i := range risks {
		for j := range len(risks) - i - 1 {
			if risks[j].Severity < risks[j+1].Severity {
				risks[j], risks[j+1] = risks[j+1], risks[j]
			}
		}
	}

	if len(risks) == 0 {
		return []Risk{NoRisk}
	} else if risks[0] == TakeOverCluster {
		return []Risk{TakeOverCluster}
	}
	return risks
}

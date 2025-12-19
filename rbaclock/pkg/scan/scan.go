package scan

import (
	cli "rbaclock/pkg/client"
	priv "rbaclock/pkg/privilege"
	"strings"
)

type Alert struct {
	Namespace      string
	ServiceAccount string
	Risks          []priv.Risk
	Privileges     []priv.Privilege
}

var ExcludeSA = []string{}

func Scan() map[string][]Alert {
	saPrivs := map[string][]Priv{}
	saEffects := map[string][]priv.Risk{}
	// only check sa exists on worker nodes
	nodes := cli.WorkerNodes()
	for _, node := range nodes {
		nodeSaPrivs := TakeNodeSaPrivs(node)
		for k, v := range nodeSaPrivs {
			saPrivs[k] = v
		}
	}

	for sa, privs := range saPrivs {
		ExcludeSA = []string{}
		effects := EvilDeduction(privs, sa)
		// if len(effects) == 1 && effects[0] == priv.NoRisk {
		// 	continue
		// }
		saEffects[sa] = effects
	}

	nsAlerts := map[string][]Alert{}
	for index, effects := range saEffects {
		ns := strings.Split(index, ":")[0]
		sa := strings.Split(index, ":")[1]

		alert := Alert{
			Namespace:      ns,
			ServiceAccount: sa,
			Risks:          effects,
			Privileges:     priv.PrivsOutput(saPrivs[index]),
		}

		if _, ok := nsAlerts[ns]; !ok {
			nsAlerts[ns] = []Alert{}
		}
		nsAlerts[ns] = append(nsAlerts[ns], alert)
	}
	return nsAlerts
}

func EvilDeduction(privs []Priv, sa string) []priv.Risk {
	if inArray(sa, ExcludeSA) {
		return []priv.Risk{}
	} else {
		ExcludeSA = append(ExcludeSA, sa)
	}
	result := []priv.Risk{}
	risks := priv.PrivRisks(privs)
	result = priv.MergeRisk(result, risks)
	saPrivs := map[string][]Priv{}
	for _, r := range risks {
		switch r {
		// case priv.TakeNamespacePods:
		// 	saPrivs = TakeNamespacePodPrivs(namespace(sa))
		// case priv.TakeClusterSaToken:
		// 	saPrivs = TakeClusterSaTokenPrivs()
		case priv.TakeOverCluster:
			return []priv.Risk{priv.TakeOverCluster}
		case priv.TakeOverNode:
			saPrivs = TakeOverNodeSaPrivs()
			result = priv.MergeRisk(result, []priv.Risk{priv.TakeOverNode})
		case priv.TakeOverPod:
			// take over namespace pod
			saPrivs = TakeNamespaceSaPrivs(namespace(sa))
			result = priv.MergeRisk(result, priv.TakeOverPod)
		case priv.DestroyCurrentNode:
			result = priv.MergeRisk(result, priv.DestroyCurrentNode)
		case priv.DestroyClusterContainer:
			result = priv.MergeRisk(result, priv.DestroyClusterContainer)
		case priv.DestroyNamespaceContainer:
			result = priv.MergeRisk(result, priv.DestroyNamespaceContainer)
		case priv.LeakInfo:
			result = priv.MergeRisk(result, priv.LeakInfo)
		// case priv.PrivEscalation:
		// 	result = priv.MergeRisk(result, priv.PrivEscalation)
		// 	privs := TakeNamespaceRolePriv(namespace(sa))
		// 	// pop the service account for re-evaluation
		// 	ExcludeSA = ExcludeSA[:len(ExcludeSA)-1]
		// 	result = priv.MergeRisk(result, EvilDeduction(privs, sa))
		case priv.TakeNamespaceRoles:
			privs := TakeNamespaceRolePrivs(namespace(sa))
			// pop the service account for re-evaluation
			ExcludeSA = ExcludeSA[:len(ExcludeSA)-1]
			result = priv.MergeRisk(result, EvilDeduction(privs, sa))
		case priv.TakeNamespaceSa:
			saPrivs = TakeNamespaceSaPrivs(namespace(sa))
		case priv.TakeNamespaceSaTokens:
			saPrivs = TakeNamespaceSaTokenPrivs(namespace(sa))
		case priv.NoRisk:
			continue
		}

		if len(saPrivs) > 0 {
			for sa, privs := range saPrivs {
				newRisks := EvilDeduction(privs, sa)
				if len(newRisks) > 0 && newRisks[0] == priv.TakeOverCluster {
					return []priv.Risk{priv.TakeOverCluster}
				}
				result = priv.MergeRisk(result, newRisks)
			}
		}
	}
	return result
}

func namespace(sa string) string {
	return strings.Split(sa, ":")[0]
}

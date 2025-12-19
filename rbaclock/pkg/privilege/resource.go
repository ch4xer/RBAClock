package privilege

import (
	"rbaclock/pkg/db"
	"slices"
	"strings"
)

var criticalResources = map[string]bool{
	"*":                               true,
	"pods":                            true,
	"services":                        true,
	"deployments":                     true,
	"replicasets":                     true,
	"secrets":                         true,
	"nodes":                           true,
	"roles":                           true,
	"rolebindings":                    true,
	"clusterroles":                    true,
	"ingresses":                       true,
	"configmaps":                      true,
	"networkpolicies":                 true,
	"clusterrolebindings":             true,
	"certificatesigningrequests":      true,
	"serviceaccounts":                 true,
	"statefulsets":                    true,
	"jobs":                            true,
	"daemonsets":                      true,
	"cronjobs":                        true,
	"users":                           true,
	"groups":                          true,
	"mutatingwebhookconfigurations":   true,
	"validatingwebhookconfigurations": true,
}

func IsCriticalBuiltin(resource string) bool {
	r := strings.Split(resource, "/")[0]
	return criticalResources[r]
}

func isCriticalCR(resource string) bool {
	builtins := db.QueryCRMap(resource)
    return slices.ContainsFunc(builtins, func(builtin string) bool {
        return IsCriticalBuiltin(builtin)
    })
}

func FilterCriticalBuiltin(resources []string) []string {
	var res []string
	for _, resource := range resources {
		if IsCriticalBuiltin(resource) {
			res = append(res, resource)
		}
	}
	return res
}

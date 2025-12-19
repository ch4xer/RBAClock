package privilege

import (
	"fmt"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
)

// * there is an assumption
// * that the privileges of one sa can only affect
// * one namespace if it is not a cluster-scoped sa
func isSubsetPrivs(subset, privs []Privilege) bool {
	privMap := map[Privilege]bool{}
	for _, p := range privs {
		privMap[p] = true
	}
	for _, p := range subset {
		patternParts := strings.Split(string(p), ":")
		verbs := strings.Split(patternParts[2], ",")
		satisified := false
		for _, verb := range verbs {
			if patternParts[0] == "*" {
				i1 := Privilege("cluster:" + patternParts[1] + ":" + verb)
				i2 := Privilege("namespace:" + patternParts[1] + ":" + verb)
				i3 := Privilege("cluster:*:" + verb)
				i4 := Privilege("namespace:*:" + verb)
				i5 := Privilege("cluster:" + patternParts[1] + ":*")
				i6 := Privilege("namespace:" + patternParts[1] + ":*")
				i7 := Privilege("cluster:*:*")
				i8 := Privilege("namespace:*:*")

				if privMap[i1] ||
					privMap[i2] ||
					privMap[i3] ||
					privMap[i4] ||
					privMap[i5] ||
					privMap[i6] ||
					privMap[i7] ||
					privMap[i8] {
					satisified = true
					break
				}
			} else {
				i1 := Privilege(patternParts[0] + ":" + patternParts[1] + ":" + verb)
				i2 := Privilege(patternParts[0] + ":*:" + verb)
				i3 := Privilege(patternParts[0] + patternParts[1] + ":*")
				i4 := Privilege(patternParts[0] + ":*:*")
				if privMap[i1] ||
					privMap[i2] ||
					privMap[i3] ||
					privMap[i4] {
					satisified = true
					break
				}
			}
		}
		if !satisified {
			return false
		}
	}
	return true
}

func MergeRisk(arr1 []Risk, arr2 any) []Risk {
	result := []Risk{}
	effectMap := map[Risk]bool{}
	for _, e := range arr1 {
		effectMap[e] = true
	}
	switch a := arr2.(type) {
	case []Risk:
		for _, e := range a {
			effectMap[e] = true
		}
	case Risk:
		effectMap[a] = true
	}
	for k := range effectMap {
		result = append(result, k)
	}
	return Process(result)
}

func PrivEntry(rules []rbacv1.PolicyRule, namespace string) []Privilege {
	result := []Privilege{}
	var prefix string
	if namespace == "" {
		prefix = "cluster"
	} else {
		prefix = namespace
	}
	for _, rule := range rules {
		for _, resource := range rule.Resources {
			// * did not check policy with resourceNames, can be optimized
			if len(rule.ResourceNames) > 0 {
				continue
			}
			// NOTE: need to run `rbaclock crcheck` to set up the CRMap
			if !IsCriticalBuiltin(resource) || !isCriticalCR(resource) {
				continue
			}

			for _, verb := range rule.Verbs {
				result = append(result, Privilege(fmt.Sprintf("%s:%s:%s", prefix, resource, verb)))
			}
		}
	}
	return result
}

func MergePrivs(p1, p2 []Privilege) []Privilege {
	privMap := map[Privilege]bool{}
	for _, p := range p1 {
		privMap[p] = true
	}
	for _, p := range p2 {
		privMap[p] = true
	}
	result := []Privilege{}
	for k := range privMap {
		result = append(result, k)
	}
	return result
}

func RisksSubstract(set1, set2 []Risk) []Risk {
	riskMap := map[Risk]bool{}
	for _, r := range set2 {
		riskMap[r] = true
	}
	result := []Risk{}
	for _, r := range set1 {
		if !riskMap[r] {
			result = append(result, r)
		}
	}
	return Process(result)
}

func RiskScore(risks []Risk) int {
	score := 0
	for _, r := range risks {
		score += int(r.Severity)
	}
	return score
}

func PrivsOutput(privs []Privilege) []Privilege {
	data := map[string]map[string][]string{}
	for _, priv := range privs {
		parts := strings.Split(string(priv), ":")
		ns := parts[0]
		resource := parts[1]
		verb := parts[2]
		if _, ok := data[ns]; !ok {
			data[ns] = map[string][]string{}
		}
		if _, ok := data[ns][resource]; !ok {
			data[ns][resource] = []string{}
		}
		data[ns][resource] = append(data[ns][resource], verb)
	}
	output := []Privilege{}
	for ns := range data {
		for resource := range data[ns] {
			output = append(output, Privilege(fmt.Sprintf("%s:%s:%s", ns, resource, strings.Join(data[ns][resource], ","))))
		}
	}

	return output
}

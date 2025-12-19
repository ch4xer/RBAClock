package measure

import (
	"rbaclock/conf"
	priv "rbaclock/pkg/privilege"
)

func weightMap(initVec map[string][]float32) map[string]float32 {
	weight := map[string]float32{}
	for o := range initVec {
		if nameIsPod(o) {
			weight[o] = conf.Weight.Container
		} else {
			weight[o] = conf.Weight.Node
		}
	}
	return weight
}

func riskMap(risks []priv.Risk, namespace, node string, initVec map[string][]float32) map[string][]float32 {
	riskMap := copyMap(initVec)
	for _, r := range risks {
		switch r {
		case priv.TakeOverCluster:
			for n := range riskMap {
				riskMap[n] = []float32{1, 1, 1}
			}
		case priv.TakeOverNode:
			for n := range riskMap {
				if !nameInControlPlane(n) {
					riskMap[n] = []float32{1, 1, 1}
				}
			}
		case priv.TakeOverPod:
			for n := range riskMap {
				// only exec into the pod
				if !nameInControlPlane(n) && nameIsPod(n) {
					riskMap[n] = []float32{0, 0, 1}
				}
			}
		case priv.DestroyCurrentNode:
			// can only modify the current node
			for n := range riskMap {
				if n == node {
					riskMap[n] = []float32{0, 1, 0}
				}
			}
		case priv.DestroyClusterContainer:
			for n := range riskMap {
				if nameIsPod(n) {
					riskMap[n] = []float32{0, 1, 0}
				}
			}
		case priv.DestroyNamespaceContainer:
			for n := range riskMap {
				if nameIsPod(n) && nameInNamespace(n, namespace) {
					riskMap[n] = []float32{0, 1, 0}
				}
			}
		case priv.LeakInfo:
			for n := range riskMap {
				if nameIsPod(n) && nameInNamespace(n, namespace) {
					riskMap[n] = []float32{1, 0, 0}
				}
			}
		}
	}

	return riskMap
}

func copyMap(vec map[string][]float32) map[string][]float32 {
	newVec := map[string][]float32{}
	for k, v := range vec {
		newSlice := make([]float32, len(v))
		copy(newSlice, v)
		newVec[k] = newSlice
	}
	return newVec
}

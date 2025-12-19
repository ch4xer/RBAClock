package measure

import (
	"log"
	cli "rbaclock/pkg/client"
	"rbaclock/pkg/db"
	"rbaclock/pkg/scan"
	"strings"

	v1 "k8s.io/api/core/v1"
)

var (
	controlPlaneNodes []*v1.Node
	controlPlanePods  []*v1.Pod
)

func RecordRiskVec() {
	initMap := map[string][]float32{}
	controlPlaneNodes = cli.ControlPlaneNodes()
	controlPlanePods = cli.PodsInControlPlane()
	nodes := cli.Nodes()
	for _, node := range nodes {
		// read, write(modify, disrupt), exec
		initMap[node.Name] = []float32{0, 0, 0}
		pods := cli.PodsOnNode(node)
		for _, pod := range pods {
			index := pod.Namespace + "/" + pod.Name
			if nameInControlPlane(pod.Spec.NodeName) {
				controlPlanePods = append(controlPlanePods, pod)
			}
			initMap[index] = []float32{0, 0, 0}
		}
	}
	db.InitVectorTable(len(initMap) * 3)

	wp := weightMap(initMap)

	// get all risks proposed by pods on worker nodes
	// and convert them into vectors
	nsAlerts := scan.Scan()
	log.Println("Risk scan finished, converting to vectors...")
	for ns, alerts := range nsAlerts {
		for _, alert := range alerts {
			sa := alert.ServiceAccount
			pods := cli.PodsBoundToSa(ns, sa)
			for _, pod := range pods {
				if !nameInControlPlane(pod.Name) {
					n := pod.Spec.NodeName
					rp := riskMap(alert.Risks, ns, n, initMap)
					riskVec := risk2Vec(rp, wp)
					db.UpsertPodVec(pod.Name, sa, ns, n, riskVec)
				}
			}
		}
	}
	log.Println("Recording risk vectors finished.")
}

// func MeasureNode(subClusterSize int) (float32, []*cluster.Cluster) {
//
// 	clusterNum := cluster.ClusterNum(subClusterSize)
// 	clusters := cluster.ClusterNodes(clusterNum)
// 	car := cluster.CAR(clusters, "node")
// 	l1 := db.VectorL1(car)
// 	return l1, clusters
// }

// measure erp for cluster
func MeasureCluster() float32 {
	result := float32(0)
	nodes := cli.WorkerNodes()
	for _, node := range nodes {
		result += MeasureNode(node)
	}
	return result
}

// measure the erp for each node
func MeasureNode(node *v1.Node) float32 {
	result := float32(0)
	nodeRiskVec := db.QueryNodeVec(node.Name)
	pods := cli.PodsOnNode(node)
	for _, pod := range pods {
		podRiskVec := db.QuerySAVec(pod.Spec.ServiceAccountName)
		result += db.VectorL1(db.XORVector(nodeRiskVec, podRiskVec))
	}
	return result
}

// func mergeVecInGroup(groupVecs map[int]map[string][]float32) map[int][]float32 {
// 	mergedVec := map[int][]float32{}
// 	for group, vecMap := range groupVecs {
// 		for _, vec := range vecMap {
// 			if len(mergedVec[group]) == 0 {
// 				mergedVec[group] = vec
// 			} else {
// 				mergedVec[group] = db.MergeVector(mergedVec[group], vec)
// 			}
// 		}
// 	}
// 	return mergedVec
// }

func nameInNamespace(name, namespace string) bool {
	ns := strings.Split(name, "/")[0]
	return ns == namespace
}

func nameIsPod(name string) bool {
	return strings.Contains(name, "/")
}

func nameInControlPlane(name string) bool {
	for _, node := range controlPlaneNodes {
		if name == node.Name {
			return true
		}
	}
	for _, pod := range controlPlanePods {
		if name == pod.Name {
			return true
		}
	}
	return false
}

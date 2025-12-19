package emulate

import (
	"log"
	cli "rbaclock/pkg/client"
	"rbaclock/pkg/db"

	v1 "k8s.io/api/core/v1"
)

func Emulate() {
	clusterRisk := clusterRisk()
	escalationTimes := clusterEscalationTimes()
	dangerNodeRate := dangerNodeRate()
	log.Printf("Cluster Risk: %f\n", clusterRisk)
	log.Printf("Cluster Escalation Times: %d\n", escalationTimes)
	log.Printf("Danger Node Rate: %f\n", dangerNodeRate)
}

// count the added risk of each node
func clusterRisk() float32 {
	nodes := cli.WorkerNodes()
	risk := float32(0)
	for _, node := range nodes {
		risk += nodeRisk(node)
	}
	return risk
}

// count the union risk of each pod for each node
func nodeRisk(node *v1.Node) float32 {
	nodeVec := db.QueryNodeVec(node.Name)
	return db.VectorL1(nodeVec)
}

// count the times of potential escalation for cluster
func clusterEscalationTimes() int {
	nodes := cli.WorkerNodes()
	count := 0
	for _, node := range nodes {
		count += nodeEscalationTimes(node)
	}
	return count
}

// count the times of potential escalation for each node
func nodeEscalationTimes(node *v1.Node) int {
	nodeVec := db.QueryNodeVec(node.Name)
	pods := cli.PodsOnNode(node)
	count := 0
	for _, pod := range pods {
		podVec := db.QueryPodVec(pod.Name)
		if db.VectorL1(db.XORVector(nodeVec, podVec)) > 0 {
			count++
		}
	}
	return count
}

// count the number of nodes that can take over cluster
func dangerNodeRate() float32 {
	nodes := cli.WorkerNodes()
	dangerNode := []string{}
	for _, node := range nodes {
		nodeVec := db.QueryNodeVec(node.Name)
		isDanger := true
		if len(nodeVec) > 0 {
			for _, v := range nodeVec {
				if int(v) == 0 {
					isDanger = false
					break
				}
			}
		} else {
			isDanger = false
		}
		if isDanger {
			dangerNode = append(dangerNode, node.Name)
		}
	}
	log.Printf("danger node: %v\n", dangerNode)
	dangerRate := float32(len(dangerNode)) / float32(len(nodes))
	return dangerRate
}

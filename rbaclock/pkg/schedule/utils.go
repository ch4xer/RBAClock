package schedule

import (
	"rbaclock/pkg/group"

	v1 "k8s.io/api/core/v1"
)

// func addPod2Cluster(plan []*group.Group, label int, pod *v1.Pod) []*group.Group {
// 	for _, cluster := range plan {
// 		if cluster.Label == label {
// 			cluster.Pods = append(cluster.Pods, pod)
// 			return plan
// 		}
// 	}
// 	cluster := &group.Group{
// 		Label: label,
// 		Pods:  []*v1.Pod{pod},
// 		Nodes: []*v1.Node{},
// 	}
// 	plan = append(plan, cluster)
// 	return plan
// }

func addNode2Group(clusters []*group.Group, label int, node *v1.Node) []*group.Group {
	for _, cluster := range clusters {
		if cluster.Label == label {
			cluster.Nodes = append(cluster.Nodes, node)
			return clusters
		}
	}
	cluster := &group.Group{
		Label: label,
		Pods:  []*v1.Pod{},
		Nodes: []*v1.Node{node},
	}
	clusters = append(clusters, cluster)
	return clusters
}

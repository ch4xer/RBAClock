package client

import (
	"context"
	"errors"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

var ErrNodeAlreadyAnchored = errors.New("node already anchored")

func Nodes() []*v1.Node {
	nodes, _ := Client().CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	result := []*v1.Node{}
	for _, node := range nodes.Items {
		if isNodeSchedulable(&node) {
			result = append(result, &node)
		}
	}
	return result
}

func isNodeSchedulable(node *v1.Node) bool {
	if node.Spec.Unschedulable {
		return false
	}

	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
			return false
		}
	}

	return true
}

func ControlPlaneNodes() []*v1.Node {
	result := []*v1.Node{}
	nodes := Nodes()
	for _, node := range nodes {
		if isControlPlane(node) {
			result = append(result, node)
		}
	}
	return result
}

func WorkerNodes() []*v1.Node {
	result := []*v1.Node{}
	nodes := Nodes()
	for _, node := range nodes {
		if isControlPlane(node) {
			continue
		}
		result = append(result, node)
	}
	return result
}

func isControlPlane(node *v1.Node) bool {
	if _, hasMasterLabel := node.Labels["node-role.kubernetes.io/master"]; hasMasterLabel {
		return true
	}
	if _, hasControlPlaneLabel := node.Labels["node-role.kubernetes.io/control-plane"]; hasControlPlaneLabel {
		return true
	}
	return false
}

func CheckTaint(node *v1.Node, keyword string) bool {
	for _, taint := range node.Spec.Taints {
		taintStr := taint.Key + "=" + taint.Value
		if strings.Contains(taintStr, keyword) {
			return true
		}
	}
	return false
}

func UntaintAllWorkerNodes(key string) {
	nodes := WorkerNodes()
	for _, node := range nodes {
		taints := []v1.Taint{}
		for _, taint := range node.Spec.Taints {
			if taint.Key != key {
				taints = append(taints, taint)
			}
		}
		node.Spec.Taints = taints
		_, _ = Client().CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	}
}

func UnlabelAllWorkerNodes(keyword string) {
	nodes := WorkerNodes()
	for _, node := range nodes {
		labels := map[string]string{}
		for k, v := range node.Labels {
			if !strings.Contains(k, keyword) {
				labels[k] = v
			}
		}
		node.Labels = labels
		_, _ = Client().CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	}
}

type NodeInfo struct {
	Name   string
	PodNum int
	Info   *v1.Node
}

func ListNodesByPodNum() []NodeInfo {
	nodePodNum := map[string]int{}
	allNode := WorkerNodes()
	for _, node := range allNode {
		nodePodNum[node.Name] = 0
	}
	pods := Pods("")
	for _, pod := range pods {
		if pod.Spec.NodeName != "" {
			nodePodNum[pod.Spec.NodeName]++
		}
	}

	nps := []NodeInfo{}
	for name, podNum := range nodePodNum {
		info := Node(name)
		if isControlPlane(info) {
			continue
		}
		nps = append(nps, NodeInfo{name, podNum, info})
	}
	for i, np := range nps {
		for j := i + 1; j < len(nps); j++ {
			if nps[j].PodNum < np.PodNum {
				nps[i], nps[j] = nps[j], nps[i]
			}
		}
	}
	return nps
}

func Node(name string) *v1.Node {
	node, _ := Client().CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	return node
}

// NOTE: UpsertNodeAnchor will overide conflicts
func UpsertNodeAnchor(node *v1.Node, key, value string) error {
	taintAndLabelNode := func(node *v1.Node, key, value string) error {
		newTaint := []v1.Taint{}
		for _, taint := range node.Spec.Taints {
			// the node has already been tainted
			// if taint exists, then label must exist
			if taint.Key != key {
				newTaint = append(newTaint, taint)
			}
		}
		node.Spec.Taints = newTaint
		node.Spec.Taints = append(node.Spec.Taints, v1.Taint{
			Key:    key,
			Value:  value,
			Effect: v1.TaintEffectNoSchedule,
		})

		if node.Labels == nil {
			node.Labels = map[string]string{}
		}
		if val, ok := node.Labels[key]; !ok || val != value {
			node.Labels[key] = value
		}
		_, err := Client().CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
		return err
	}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := taintAndLabelNode(node, key, value)
		return err
	})

	return err
}

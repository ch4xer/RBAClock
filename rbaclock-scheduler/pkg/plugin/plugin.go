package plugin

import (
	"context"
	"scheduler/pkg/db"
	"scheduler/pkg/utils"
	"slices"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	log "k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const Name = "RBAClock"

var (
	ignoreNamespaces = []string{"kube-system"}
	ignoreNodes      = []string{"minikube"}
)

// RBAClock is a plugin that logs scheduling information.
type RBAClock struct{}

var _ framework.ScorePlugin = &RBAClock{}

// Name returns the name of the plugin.
func (p *RBAClock) Name() string {
	return Name
}

// Score logs the scheduling information and returns a score.
func (p *RBAClock) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	// NOTE: skip scheduling under certain conditions
	if slices.Contains(ignoreNamespaces, pod.Namespace) || slices.Contains(ignoreNodes, nodeName) {
		return 0, framework.NewStatus(framework.Success)
	}
	delta := calcDeltaERP4Node(pod, nodeName)
	score := max(100-normalizeDelta(delta), 0)
	log.Infof("Score %s/%s on %s: %d", pod.Namespace, pod.Name, nodeName, int(score))
	return int64(score), framework.NewStatus(framework.Success)
}

// PostBind logs the final scheduling decision for the pod.
func (p *RBAClock) PostBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	log.Infof("Decision: %s/%s => %s", pod.Namespace, pod.Name, nodeName)
	updateState(pod, nodeName)
}

// ScoreExtensions returns the ScoreExtensions interface.
func (p *RBAClock) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(ctx context.Context, _ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &RBAClock{}, nil
}

// normalize Delta to range from 0 to 100
func normalizeDelta(delta float32) float32 {
	return delta / 10
}

func calcDeltaERP4Node(pod *v1.Pod, node string) float32 {
	podVecs := db.QueryPodVecsOnNode(node)
	nodeVec := db.QueryNodeVec(node)
	var oldERP float32 = 0
	for _, vec := range podVecs {
		oldERP += utils.VectorL1(utils.XORVector(vec, nodeVec))
	}

	newPodVec := db.QuerySAVec(pod.Spec.ServiceAccountName)
	newNodeVec := utils.MergeVector(nodeVec, newPodVec)

	newERP := utils.VectorL1(utils.XORVector(newNodeVec, newPodVec))
	for _, vec := range podVecs {
		newERP += utils.VectorL1(utils.XORVector(vec, newNodeVec))
	}
	return newERP - oldERP
}

func updateState(pod *v1.Pod, nodeName string) {
	podVec := db.QuerySAVec(pod.Spec.ServiceAccountName)
	db.UpsertPodVec(pod.Name, pod.Spec.ServiceAccountName, pod.Namespace, nodeName, podVec)
	log.Infof("Update %s state", nodeName)
}

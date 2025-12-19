package client

import (
	"context"
	"errors"
	"strings"
	"time"

	"rbaclock/conf"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shouldIgnorePod(pod *v1.Pod) bool {
	if shouldIgnoreNamespace(pod.Namespace) {
		return true
	}
	for _, ignore := range conf.IgnorePods {
		if strings.Contains(pod.Name, ignore) {
			return true
		}
	}
	return false
}

func Pods(namespace string) []*v1.Pod {
	// FieldSelector: "status.phase=Running",
	pods, _ := Client().CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	result := []*v1.Pod{}
	for _, pod := range pods.Items {
		if shouldIgnorePod(&pod) {
			continue
		}
		result = append(result, &pod)
	}
	return result
}

func PodsOnNode(node *v1.Node) []*v1.Pod {
	pods, _ := Client().CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	result := []*v1.Pod{}
	for _, pod := range pods.Items {
		if pod.Spec.NodeName == node.Name && !shouldIgnorePod(&pod) {
			result = append(result, &pod)
		}
	}
	return result
}

func PodsInControlPlane() []*v1.Pod {
	nodes := ControlPlaneNodes()
	result := []*v1.Pod{}
	for _, node := range nodes {
		pods := PodsOnNode(node)
		for _, pod := range pods {
			result = append(result, pod)
		}
	}
	return result
}

func PodsBoundToSa(namespace, sa string) []*v1.Pod {
	pods := Pods(namespace)
	var targetPods []*v1.Pod
	for _, pod := range pods {
		if pod.Spec.ServiceAccountName == sa {
			targetPods = append(targetPods, pod)
		}
	}
	return targetPods
}

func removePodToleration(pod *v1.Pod, key string) {
	if len(pod.OwnerReferences) == 0 {
		var updatedTolerations []v1.Toleration
		for _, toleration := range pod.Spec.Tolerations {
			if toleration.Key != key {
				updatedTolerations = append(updatedTolerations, toleration)
			}
		}

		pod.Spec.Tolerations = updatedTolerations
		_, _ = Client().CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	}
}

func removePodAffinity(pod *v1.Pod, key string) {
	if len(pod.OwnerReferences) == 0 {
		if pod.Spec.Affinity != nil {
			if pod.Spec.Affinity.NodeAffinity != nil {
				if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
					newNodeSelectorTerms := []v1.NodeSelectorTerm{}
					for _, term := range pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
						newMatchExpressions := []v1.NodeSelectorRequirement{}
						for _, expr := range term.MatchExpressions {
							if expr.Key != key {
								newMatchExpressions = append(newMatchExpressions, expr)
							}
						}
						term.MatchExpressions = newMatchExpressions
						newNodeSelectorTerms = append(newNodeSelectorTerms, term)
					}
					pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = newNodeSelectorTerms
				}
			}
		}
		// pod.Spec.Affinity.NodeAffinity = &v1.NodeAffinity{}
	}
}

func createPod(pod *v1.Pod) error {
	pod.ResourceVersion = ""
	pod.UID = ""
	_, err := Client().CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	return err
}

// func RecreateNamespacePods(namespace string, skipNodes []string) {
// 	pods := ListPods(namespace)
// 	for _, pod := range pods {
// 		skip := false
// 		for _, node := range skipNodes {
// 			if pod.Spec.NodeName == node {
// 				skip = true
// 				break
// 			}
// 		}
// 		if !skip {
// 			DeletePod(&pod)
// 			// if pod has controller, dont recreate it manually
// 			if len(pod.OwnerReferences) == 0 {
// 				CreatePod(&pod)
// 			}
// 		}
// 	}
// }

// func FilterNodePodsByToleration(value string) {
// 	pods, _ := Client().CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
// 	targetPods := []v1.Pod{}
// 	for _, pod := range pods.Items {
// 		if pod.Spec.NodeName == value && !shouldIgnorePod(pod) {
// 			checked := false
// 			for _, t := range pod.Spec.Tolerations {
// 				if strings.Contains(t.Key, conf.GlobalKey) {
// 					if t.Value == value {
// 						checked = true
// 						break
// 					}
// 				}
// 			}
// 			if !checked {
// 				targetPods = append(targetPods, pod)
// 			}
// 		}
// 	}
// 	DeletePods(targetPods)
// 	for _, pod := range targetPods {
// 		if len(pod.OwnerReferences) == 0 {
// 			CreatePod(&pod)
// 		}
// 	}
// }

func Pod(namespace, name string) *v1.Pod {
	pod, err := Client().CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil
	}
	return pod
}

func upsertPodSpecAnchor(pod *v1.Pod, key, value string) {
	pod.Spec.NodeName = value
	pod.ObjectMeta.ResourceVersion = ""
	pod.ObjectMeta.UID = ""
	// add toleration
	haveToleration := false
	for i, toleration := range pod.Spec.Tolerations {
		if toleration.Key == key {
			pod.Spec.Tolerations[i].Value = value
			haveToleration = true
		}
	}
	if !haveToleration {
		toleration := v1.Toleration{
			Key:    key,
			Value:  value,
			Effect: v1.TaintEffectNoSchedule,
		}
		pod.Spec.Tolerations = append(pod.Spec.Tolerations, toleration)
	}

	// add affinity
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.NodeAffinity != nil {
			if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				for i, term := range pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
					for j, expr := range term.MatchExpressions {
						if expr.Key == key {
							pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions[j].Values = []string{value}
							return
						}
					}
				}
			}
		}
	}

	pod.Spec.Affinity = &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      key,
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{value},
							},
						},
					},
				},
			},
		},
	}
}

// delete pod and wait
// prevent conflict when recreate pod
func DeletePod(pod *v1.Pod) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(5 * time.Second)
	name := pod.Name
	namespace := pod.Namespace

	deletePolicy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(0)
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
		PropagationPolicy:  &deletePolicy,
	}
	err := Client().CoreV1().Pods(namespace).Delete(context.TODO(), name, deleteOptions)

	if err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			pod := Pod(namespace, name)
			if pod == nil {
				return nil
			}
		case <-timeoutChan:
			return errors.New("delete pod timeout")
		}
	}
}

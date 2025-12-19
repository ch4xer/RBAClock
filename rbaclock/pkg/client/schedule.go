package client

import (
	"context"
	"reflect"

	"slices"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// NOTE: UpsertPodAnchor will overide conflicts
func UpsertPodAnchor(pod *v1.Pod, key, value string) error {
	// add toleration to pod or owner
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var err error
		// pod may be recreated by modified owner
		if pod == nil {
			return nil
		}
		owner := Owner(pod)
		if !reflect.ValueOf(owner).IsNil() {
			// update owner will recreate pods
			upsertOwnerSpecAnchor(owner, key, value)
			err = updateOwner(owner)
		} else {
			// update pod spec, and recreate pod
			newPod := pod.DeepCopy()
			upsertPodSpecAnchor(newPod, key, value)
			err = createPod(newPod)
			if err == nil {
				err = DeletePod(pod)
			}
		}
		return err
	})

	return err
}

func RecreateIrrelavantPods(node string, whitelist []string) error {
	pods, _ := Client().CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
	})

	for _, pod := range pods.Items {
		if shouldIgnorePod(&pod) {
			continue
		}
		skip := slices.Contains(whitelist, pod.Name)
		if skip {
			continue
		}

		var err error
		ref := pod.DeepCopy()
		hasOwner := false
		if Owner(&pod) != nil {
			hasOwner = true
		}
		err = DeletePod(&pod)

		if err != nil {
			return err
		}
		// if no owner, recreate manually
		// otherwise, owner will recreate it automatically
		if !hasOwner {
			ref.Spec.NodeName = ""
			err = createPod(ref)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// func addTolerationToPodOrOwner(namespace, podName, key, value string) {
// 	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
// 		var err error
// 		pod := Pod(namespace, podName)
// 		if len(pod.OwnerReferences) > 0 {
// 			for _, owner := range pod.OwnerReferences {
// 				switch owner.Kind {
// 				case "ReplicaSet":
// 					err = addOwnerToleration(
// 						getDeploymentByPod(pod),
// 						key, value,
// 					)
// 				case "DaemonSet":
// 					err = addOwnerToleration(
// 						getDaemonSetByPod(pod),
// 						key, value,
// 					)
// 				case "StatefulSet":
// 					err = addOwnerToleration(
// 						getStatefulSetByPod(pod),
// 						key, value,
// 					)
// 				}
// 			}
// 		} else {
// 			err = addPodToleration(pod, key, value)
// 		}
// 		return err
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func addAffinityToPodOrOwner(pod *v1.Pod, key, value string) {
// 	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
// 		var err error
// 		if len(pod.OwnerReferences) > 0 {
// 			for _, owner := range pod.OwnerReferences {
// 				switch owner.Kind {
// 				case "ReplicaSet":
// 					err = addOwnerAffinity(getDeploymentByPod(pod), key, value)
// 				case "DaemonSet":
// 					err = addOwnerAffinity(getDaemonSetByPod(pod), key, value)
// 				case "StatefulSet":
// 					err = addOwnerAffinity(getStatefulSetByPod(pod), key, value)
// 				}
// 			}
// 		} else {
// 			err = addPodAffinity(pod, key, value)
// 		}
// 		return err
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// }

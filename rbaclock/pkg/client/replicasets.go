package client

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	// "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// func AddTolerationToOwners(namespace, key, value string) {
// 	// get daemonsets, deployments, statefulsets in namespace
// 	daemonsets, _ := Client().AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
// 	deployments, _ := Client().AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
// 	statefulsets, _ := Client().AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})

// 	// add tolerations to pod templates
// 	for _, daemonset := range daemonsets.Items {
// 		addTolerationToPodTemplate(&daemonset.Spec.Template, key, value)
// 		_, _ = Client().AppsV1().DaemonSets(namespace).Update(context.TODO(), &daemonset, metav1.UpdateOptions{})
// 	}
// 	for _, deployment := range deployments.Items {
// 		addTolerationToPodTemplate(&deployment.Spec.Template, key, value)
// 		_, _ = Client().AppsV1().Deployments(namespace).Update(context.TODO(), &deployment, metav1.UpdateOptions{})
// 	}
// 	for _, statefulset := range statefulsets.Items {
// 		addTolerationToPodTemplate(&statefulset.Spec.Template, key, value)
// 		_, _ = Client().AppsV1().StatefulSets(namespace).Update(context.TODO(), &statefulset, metav1.UpdateOptions{})
// 	}
// }

func upsertSpecTemplateToleration(template *v1.PodTemplateSpec, key, value string) {
	// if exists, replace it
	for i, toleration := range template.Spec.Tolerations {
		if toleration.Key == key {
			template.Spec.Tolerations[i].Value = value
			return
		}
	}
	template.Spec.Tolerations = append(template.Spec.Tolerations, v1.Toleration{
		Key:    key,
		Value:  value,
		Effect: v1.TaintEffectNoSchedule,
	})
}

func upsertSpecTemplateAffinity(template *v1.PodTemplateSpec, key, value string) {
	if template.Spec.Affinity != nil {
		if template.Spec.Affinity.NodeAffinity != nil {
			if template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				for i, term := range template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
					for j, requirement := range term.MatchExpressions {
						if requirement.Key == key {
							template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions[j].Values = []string{value}
							return
						}
					}
				}
			}
		}
	}
	if template.Spec.Affinity == nil {
		template.Spec.Affinity = &v1.Affinity{}
	}
	template.Spec.Affinity.NodeAffinity = &v1.NodeAffinity{
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
	}
}

func Owner(pod *v1.Pod) runtime.Object {
	if len(pod.OwnerReferences) == 0 {
		return nil
	}
	for _, owner := range pod.OwnerReferences {
		switch owner.Kind {
		case "ReplicaSet":
			return getDeploymentByPod(pod)
		case "DaemonSet":
			return getDaemonSetByPod(pod)
		case "StatefulSet":
			return getStatefulSetByPod(pod)
		}
	}
	return nil
}

func removeOwnerToleration(owner runtime.Object, key string) {
	switch o := owner.(type) {
	case *appv1.Deployment:
		removeTemplateToleration(&o.Spec.Template, key)
		_, _ = Client().AppsV1().Deployments(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
	case *appv1.DaemonSet:
		removeTemplateToleration(&o.Spec.Template, key)
		_, _ = Client().AppsV1().DaemonSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
	case *appv1.StatefulSet:
		removeTemplateToleration(&o.Spec.Template, key)
		_, _ = Client().AppsV1().StatefulSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
	}
}

func removeOwnerAffinity(owner runtime.Object, key string) {
	switch o := owner.(type) {
	case *appv1.Deployment:
		removeTemplateAffinity(&o.Spec.Template, key)
		_, _ = Client().AppsV1().Deployments(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
	case *appv1.DaemonSet:
		removeTemplateAffinity(&o.Spec.Template, key)
		_, _ = Client().AppsV1().DaemonSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
	case *appv1.StatefulSet:
		removeTemplateAffinity(&o.Spec.Template, key)
		_, _ = Client().AppsV1().StatefulSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
	}

}

func removeTemplateToleration(template *v1.PodTemplateSpec, key string) {
	var updatedTolerations []v1.Toleration
	for _, toleration := range template.Spec.Tolerations {
		if toleration.Key != key {
			updatedTolerations = append(updatedTolerations, toleration)
		}
	}
	template.Spec.Tolerations = updatedTolerations
}

func removeTemplateAffinity(template *v1.PodTemplateSpec, key string) {
	if template.Spec.Affinity != nil {
		if template.Spec.Affinity.NodeAffinity != nil {
			if template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				newNodeSelectorTerms := []v1.NodeSelectorTerm{}
				for _, term := range template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
					newMatchExpressions := []v1.NodeSelectorRequirement{}
					for _, expr := range term.MatchExpressions {
						if expr.Key != key {
							newMatchExpressions = append(newMatchExpressions, expr)
						}
					}
					term.MatchExpressions = newMatchExpressions
					newNodeSelectorTerms = append(newNodeSelectorTerms, term)
				}
				template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = newNodeSelectorTerms

			}
		}
	}
	return
	// template.Spec.Affinity.NodeAffinity = &v1.NodeAffinity{}
}

func getDeploymentByPod(pod *v1.Pod) *appv1.Deployment {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "ReplicaSet" {
			r, _ := Client().AppsV1().ReplicaSets(pod.Namespace).Get(context.TODO(), owner.Name, metav1.GetOptions{})

			for _, owner := range r.OwnerReferences {
				if owner.Kind == "Deployment" {
					d, _ := Client().AppsV1().Deployments(pod.Namespace).Get(context.TODO(), owner.Name, metav1.GetOptions{})
					return d
				}
			}
		}
	}
	return nil
}

func getStatefulSetByPod(pod *v1.Pod) *appv1.StatefulSet {
	// get statefulset by owner reference
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "StatefulSet" {
			ss, _ := Client().AppsV1().StatefulSets(pod.Namespace).Get(context.TODO(), owner.Name, metav1.GetOptions{})
			return ss
		}
	}
	return nil
}

func getDaemonSetByPod(pod *v1.Pod) *appv1.DaemonSet {
	// get daemonset by owner reference
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			ds, _ := Client().AppsV1().DaemonSets(pod.Namespace).Get(context.TODO(), owner.Name, metav1.GetOptions{})
			return ds
		}
	}
	return nil
}

// func addOwnerToleration(owner runtime.Object, key, value string) error {
// 	var err error
// 	switch o := owner.(type) {
// 	case *appv1.Deployment:
// 		addPodTemplateToleration(&o.Spec.Template, key, value)
// 		_, err = Client().AppsV1().Deployments(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
// 	case *appv1.DaemonSet:
// 		addPodTemplateToleration(&o.Spec.Template, key, value)
// 		_, err = Client().AppsV1().DaemonSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
// 	case *appv1.StatefulSet:
// 		addPodTemplateToleration(&o.Spec.Template, key, value)
// 		_, err = Client().AppsV1().StatefulSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
// 	}
// 	return err
// }

// func addOwnerAffinity(owner runtime.Object, key, value string) error {
// 	var err error
// 	switch o := owner.(type) {
// 	case *appv1.Deployment:
// 		addPodTemplateAffinity(&o.Spec.Template, key, value)
// 		_, err = Client().AppsV1().Deployments(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
// 	case *appv1.DaemonSet:
// 		addPodTemplateAffinity(&o.Spec.Template, key, value)
// 		_, err = Client().AppsV1().DaemonSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
// 	case *appv1.StatefulSet:
// 		addPodTemplateAffinity(&o.Spec.Template, key, value)
// 		_, err = Client().AppsV1().StatefulSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
// 	}
// 	return err
// }

func upsertOwnerSpecAffinity(owner runtime.Object, key, value string) {
	switch o := owner.(type) {
	case *appv1.Deployment:
		upsertSpecTemplateAffinity(&o.Spec.Template, key, value)
	case *appv1.DaemonSet:
		upsertSpecTemplateAffinity(&o.Spec.Template, key, value)
	case *appv1.StatefulSet:
		upsertSpecTemplateAffinity(&o.Spec.Template, key, value)
	}
}

func upsertOwnerSpecToleration(owner runtime.Object, key, value string) {
	switch o := owner.(type) {
	case *appv1.Deployment:
		upsertSpecTemplateToleration(&o.Spec.Template, key, value)
	case *appv1.DaemonSet:
		upsertSpecTemplateToleration(&o.Spec.Template, key, value)
	case *appv1.StatefulSet:
		upsertSpecTemplateToleration(&o.Spec.Template, key, value)
	}
}

func upsertOwnerSpecAnchor(owner runtime.Object, key, value string) {
	upsertOwnerSpecToleration(owner, key, value)
	upsertOwnerSpecAffinity(owner, key, value)
}

func updateOwner(owner runtime.Object) error {
	var err error
	switch o := owner.(type) {
	case *appv1.Deployment:
		// delete its replicasets
		replicaSets, _ := Client().AppsV1().ReplicaSets(o.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.Set(o.Spec.Selector.MatchLabels).String(),
		})
		for _, rs := range replicaSets.Items {
			defer Client().AppsV1().ReplicaSets(o.Namespace).Delete(context.TODO(), rs.Name, metav1.DeleteOptions{
				GracePeriodSeconds: func(i int64) *int64 { return &i }(5),
			})
		}
		_, err = Client().AppsV1().Deployments(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
	case *appv1.DaemonSet:
		_, err = Client().AppsV1().DaemonSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
	case *appv1.StatefulSet:
		_, err = Client().AppsV1().StatefulSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
	}
	return err
}

package main

import (
	"context"
	"github.com/dapr/dapr/cmd/cni-node/offmesh"
	"github.com/dapr/dapr/pkg/injector/annotations"
	"github.com/dapr/dapr/pkg/injector/sidecar"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"log"
)

func IsWorkerPod(pod *corev1.Pod) bool {
	return offmesh.NodeType(pod.Spec.NodeName, offmeshCluster) == offmesh.CPUNode &&
		sidecar.Annotations(pod.Annotations).GetBoolOrDefault(annotations.KeyEnabled, false)
}
func IsDaprPod(pod *corev1.Pod) bool {
	return offmesh.NodeType(pod.Spec.NodeName, offmeshCluster) == offmesh.DPUNode &&
		sidecar.Annotations(pod.Annotations).GetBoolOrDefault(sidecar.OffmeshSidecar, false)
}

func EventHandler() *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Printf("[OnAdd] pod name: %s, NodeName:%s, dapr.io/enabled: %s, offmesh/is-sidecar:%s \n", pod.ObjectMeta.Name, pod.Spec.NodeName, sidecar.Annotations(pod.Annotations)[annotations.KeyEnabled], sidecar.Annotations(pod.Annotations)[sidecar.OffmeshSidecar])
			if pod.Spec.NodeName != "" && IsWorkerPod(pod) {
				log.Println("[HandlePodAdd] handling dapr worker pod, name: ", pod.ObjectMeta.Name)
				StartSidecarPod(pod)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*corev1.Pod)
			newPod := newObj.(*corev1.Pod)
			log.Printf("[OnUpdate] pod name: %s \n", oldPod.ObjectMeta.Name)
			if oldPod.Spec.NodeName == "" && newPod.Spec.NodeName != "" && IsWorkerPod(newPod) {
				log.Printf("[OnUpdate] HandlePodAdd \n")
				StartSidecarPod(newPod)
			} else if oldPod.Status.Phase != corev1.PodRunning &&
				newPod.Status.Phase == corev1.PodRunning && IsWorkerPod(newPod) {
				daprPod, _ := kubeClient.CoreV1().Pods(newPod.ObjectMeta.Namespace).Get(context.Background(), GetSidecarPodName(newPod.ObjectMeta.Name), metav1.GetOptions{})
				if daprPod.Status.PodIP != "" {
					AddNetworkRulesToPod(newPod, daprPod.Status.PodIP)
				}
			} else if IsDaprPod(newPod) && oldPod.Status.PodIP == "" && newPod.Status.PodIP != "" {
				workerPod, _ := kubeClient.CoreV1().Pods(newPod.ObjectMeta.Namespace).Get(context.Background(), GetWorkerPodName(newPod), metav1.GetOptions{})
				if newPod.Status.Phase == corev1.PodRunning {
					AddNetworkRulesToPod(workerPod, newPod.Status.PodIP)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Println("[OnDelete] pod name: ", pod.ObjectMeta.Name)
			if IsWorkerPod(pod) {
				DeleteSidecarPod(pod)
			}
		},
	}
}

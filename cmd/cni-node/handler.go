package main

import (
	"github.com/dapr/dapr/cmd/cni-node/offmesh"
	"github.com/dapr/dapr/pkg/injector/annotations"
	"github.com/dapr/dapr/pkg/injector/sidecar"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"log"
)

func HandlePodAdd(pod *corev1.Pod) {
	if offmesh.NodeType(pod.Spec.NodeName, offmeshCluster) == offmesh.CPUNode &&
		sidecar.Annotations(pod.Annotations).GetBoolOrDefault(annotations.KeyEnabled, false) {
		log.Println("[HandlePodAdd] handling dapr worker pod, name: ", pod.ObjectMeta.Name)
		StartSidecarPod(pod)
	} else if offmesh.NodeType(pod.Spec.NodeName, offmeshCluster) == offmesh.DPUNode &&
		sidecar.Annotations(pod.Annotations).GetBoolOrDefault(sidecar.OffmeshSidecar, false) {
		log.Println("[HandlePodAdd] handling dapr sidecar pod, name: ", pod.ObjectMeta.Name)
		AddNetworkRulesToPod(pod)
	}
}

func HandlePodDel(pod *corev1.Pod) {
	if offmesh.NodeType(pod.Spec.NodeName, offmeshCluster) == offmesh.CPUNode &&
		sidecar.Annotations(pod.Annotations).GetBoolOrDefault(annotations.KeyEnabled, false) {
		log.Println("[HandlePodDel] pod name: ", pod.ObjectMeta.Name)
		DeleteSidecarPod(pod)
	}
}

func EventHandler() *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Printf("[OnAdd] pod name: %s, NodeName:%s, dapr.io/enabled: %s, offmesh/is-sidecar:%s \n", pod.ObjectMeta.Name, pod.Spec.NodeName, sidecar.Annotations(pod.Annotations)[annotations.KeyEnabled], sidecar.Annotations(pod.Annotations)[sidecar.OffmeshSidecar])
			if pod.Spec.NodeName == "" {
				log.Println("[OnAdd] no nodeName")
			}
			HandlePodAdd(pod)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*corev1.Pod)
			newPod := newObj.(*corev1.Pod)
			log.Printf("[OnUpdate] pod name: %s \n", oldPod.ObjectMeta.Name)
			if oldPod.Spec.NodeName == "" {
				log.Printf("[OnUpdate] HandlePodAdd \n")
				HandlePodAdd(newPod)
			} else if oldPod.Spec.NodeName != newPod.Spec.NodeName {
				log.Printf("[OnUpdate] HandlePodDel HandlePodAdd \n")
				HandlePodDel(oldPod)
				HandlePodAdd(newPod)
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Println("[OnDelete] pod name: ", pod.ObjectMeta.Name)
			HandlePodDel(pod)
		},
	}
}

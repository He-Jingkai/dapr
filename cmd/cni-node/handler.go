package main

import (
	"github.com/dapr/dapr/cmd/cni-node/offmesh"
	"github.com/dapr/dapr/pkg/injector/annotations"
	"github.com/dapr/dapr/pkg/injector/sidecar"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"log"
)

func EventHandler() *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Println("[OnAdd] pod name: ", pod.ObjectMeta.Name)
			if offmesh.NodeType(pod.Spec.NodeName, offmeshCluster) == offmesh.CPUNode &&
				sidecar.Annotations(pod.Annotations).GetBoolOrDefault(annotations.KeyEnabled, false) {
				log.Println("[OnAdd] handling dapr worker pod, name: ", pod.ObjectMeta.Name)
				StartSidecarPod(pod)
			} else if offmesh.NodeType(pod.Spec.NodeName, offmeshCluster) == offmesh.DPUNode &&
				sidecar.Annotations(pod.Annotations).GetBoolOrDefault(sidecar.OffmeshSidecar, false) {
				log.Println("[OnAdd] handling dapr sidecar pod, name: ", pod.ObjectMeta.Name)
				AddNetworkRulesToPod(pod)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*corev1.Pod)
			log.Printf("[OnUpdate] pod name: %s \n, oldObj:%s \n, newObj: %s\n", pod.ObjectMeta.Name, oldObj.(*corev1.Pod).String(), newObj.(*corev1.Pod).String())
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Println("[OnDelete] pod name: ", pod.ObjectMeta.Name)
			if offmesh.NodeType(pod.Spec.NodeName, offmeshCluster) == offmesh.CPUNode &&
				sidecar.Annotations(pod.Annotations).GetBoolOrDefault(annotations.KeyEnabled, false) {
				log.Println("[OnDelete] pod name: ", pod.ObjectMeta.Name)
				DeleteSidecarPod(pod)
			}
		},
	}
}

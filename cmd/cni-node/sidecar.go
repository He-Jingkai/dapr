package main

import (
	"context"
	"github.com/dapr/dapr/cmd/cni-node/offmesh"
	"github.com/dapr/dapr/pkg/injector/sidecar"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

func GetSidecarPodName(podName string) string {
	return podName + "-proxy"
}

func GetSidecarPod(pod *corev1.Pod) (*corev1.Pod, error) {
	ctr := corev1.Container{}
	err := ctr.Unmarshal([]byte(pod.Annotations[sidecar.OffmeshSidecarAnnotation]))
	newPod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: `v1`,
			Kind:       `Pod`,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetSidecarPodName(pod.ObjectMeta.Name),
			Namespace: pod.ObjectMeta.Namespace,
			Annotations: map[string]string{
				sidecar.OffmeshPairPodAnnotation: pod.ObjectMeta.Name,
				sidecar.OffmeshSidecar:           "true",
			},
		},
		Spec: corev1.PodSpec{
			NodeName:   offmesh.GetPairNode(pod.Spec.NodeName, offmeshCluster).Name,
			Containers: []corev1.Container{ctr},
		},
	}
	for _, mount := range ctr.VolumeMounts {
		for _, volume := range pod.Spec.Volumes {
			if mount.Name == volume.Name {
				newPod.Spec.Volumes = append(newPod.Spec.Volumes, volume)
				continue
			}
		}
	}
	return &newPod, err
}

func StartSidecarPod(pod *corev1.Pod) {
	sidecarPod, err := GetSidecarPod(pod)
	if err != nil {
		log.Println("[StartSidecarPod] get sidecarPod error: ", err)
		return
	}
	_, err = kubeClient.CoreV1().Pods(pod.ObjectMeta.Namespace).Create(context.Background(), sidecarPod, metav1.CreateOptions{})
	if err != nil {
		log.Println("[StartSidecarPod] add sidecarPod pod error: ", err)
		return
	}
}

func DeleteSidecarPod(pod *corev1.Pod) {
	err := kubeClient.CoreV1().Pods(pod.ObjectMeta.Namespace).Delete(context.Background(), GetSidecarPodName(pod.ObjectMeta.Name), metav1.DeleteOptions{})
	if err != nil {
		log.Println("[DeleteSidecarPod] delete sidecar pod error: ", err)
	}
}

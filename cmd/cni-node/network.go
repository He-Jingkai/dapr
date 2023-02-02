package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dapr/dapr/pkg/injector/sidecar"
	"github.com/dapr/dapr/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/remotecommand"
	"log"
)

const (
	IPTABLES         = "iptables"
	OUTPUT_CHAIN     = "OUTPUT"
	PREROUTING_CHAIN = "PREROUTING"
	TABLE_NAT        = "nat"
	LOCALHOST        = "127.0.0.1"
)

func ExecuteInContainer(podName string, podNamespace string, cmd []string) (string, string, error) {
	req := kubeClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(podNamespace).SubResource("exec")

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return "", "", fmt.Errorf("error adding to scheme: %v", err)
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(utils.GetConfig(), "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("error while creating Executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return "", "", fmt.Errorf("error in Stream: %v", err)
	}

	return stdout.String(), stderr.String(), nil
}

// iptables -A OUTPUT -t nat -p tcp --dst 127.0.0.1 --dport daprHttpPort -j DNAT --to-destination daprIP:daprHttpPort
// iptables -A OUTPUT -t nat -p tcp --dst 127.0.0.1 --dport daprGRPCPort -j DNAT --to-destination daprIP:daprGRPCPort
// iptables -A PREROUTING -t nat -p tcp --src daprIP -j SNAT --to-source 127.0.0.1

func AddNetworkRulesToPod(daprPod *corev1.Pod) {
	pairPodName := daprPod.ObjectMeta.Annotations[sidecar.OffmeshPairPodAnnotation]
	pairPodNamespace := daprPod.ObjectMeta.Namespace
	daprIP := daprPod.Status.PodIP
	daprHttpPort := ""
	daprGRPCPort := ""
	pairPod, err := kubeClient.CoreV1().Pods(pairPodNamespace).Get(context.Background(), pairPodName, metav1.GetOptions{})
	if err != nil {
		log.Println("[AddNetworkRulesToPod] get pair pod error: ", err)
		return
	}
	for _, env := range pairPod.Spec.Containers[0].Env {
		if env.Name == sidecar.UserContainerDaprHTTPPortName {
			daprHttpPort = env.Value
		} else if env.Name == sidecar.UserContainerDaprGRPCPortName {
			daprGRPCPort = env.Value
		}
	}
	cmds := [][]string{
		{
			"sh",
			"-c",
			IPTABLES,
			"-A", OUTPUT_CHAIN,
			"-t", TABLE_NAT,
			"-p", "tcp",
			"--dst", LOCALHOST,
			"--dport", daprHttpPort,
			"-j", "DNAT",
			"--to-destination", fmt.Sprintf("%s:%s", daprIP, daprHttpPort),
		},
		{
			"sh",
			"-c",
			IPTABLES,
			"-A", OUTPUT_CHAIN,
			"-t", TABLE_NAT,
			"-p", "tcp",
			"--dst", LOCALHOST,
			"--dport", daprGRPCPort,
			"-j", "DNAT",
			"--to-destination", fmt.Sprintf("%s:%s", daprIP, daprGRPCPort),
		},
		{
			"sh",
			"-c",
			IPTABLES,
			"-A", PREROUTING_CHAIN,
			"-t", TABLE_NAT,
			"-p", "tcp",
			"--src", daprIP,
			"-j", "SNAT",
			"--to-source", LOCALHOST,
		},
	}
	for _, cmd := range cmds {
		stdout, stderr, err := ExecuteInContainer(pairPodName, pairPodNamespace, cmd)
		if err != nil {
			log.Println("[AddNetworkRulesToPod] kubectl exec error: ", err)
			return
		}
		log.Printf("[AddNetworkRulesToPod] kubectl exec stdout: %v, stderr: %v", stdout, stderr)
	}
}

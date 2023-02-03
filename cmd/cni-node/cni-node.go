package main

import (
	"github.com/dapr/dapr/cmd/cni-node/offmesh"
	"github.com/dapr/dapr/utils"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	kubeClient     *kubernetes.Clientset
	offmeshCluster offmesh.ClusterConfig
)

func main() {
	offmeshCluster = offmesh.ReadClusterConfigYaml(offmesh.ClusterConfigYamlPath)
	kubeClient = utils.GetKubeClient()
	informer := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0).Core().V1().Pods().Informer()
	informer.AddEventHandler(EventHandler())
	stopper := make(chan struct{}, 2)
	go informer.Run(stopper)
	log.Println("watch pod started...")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	stopper <- struct{}{}
	close(stopper)
	log.Println("watch pod stopped...")
}

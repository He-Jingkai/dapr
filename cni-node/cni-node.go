package main

import (
	"github.com/dapr/dapr/cni-node/offmesh"
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
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0)
	informer := factory.Core().V1().Pods().Informer()
	_, err := informer.AddEventHandler(EventHandler())
	if err != nil {
		log.Println("AddEventHandler error:", err)
	}
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

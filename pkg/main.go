package main

import (
	"context"

	"os"
	"time"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/chaos"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/client"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	log.Info("Starting Sinth Chaos")

	ctx := context.Background()

	namespace := os.Getenv("NAMESPACE")
	labelSelector := os.Getenv("LABEL_SELECTOR")
	networkInterface := os.Getenv("NETWORK_INTERFACE")
	netemCommands := os.Getenv("NETEM_COMMANDS")
	socketPath := os.Getenv("SOCKET_PATH")
	destinationIPs := os.Getenv("DESTINATION_IPS")

	// Kubernetes client
	var kubernetes client.Kubernetes = client.NewClientsetKubernetes(getClientset())

	// Command cmd
	var command cmd.Command = cmd.NewOsBashExecCommand()

	// ContainerRuntime client
	var containerRuntime client.ContainerRuntime = client.NewContainerdContainerRuntime(command, socketPath)

	log.Info("Creating network-loss fault engine")
	networkLossFaulEngine := chaos.FaultEngine{
		FaultAction:      chaos.NewNetworkLossFaultAction(command),
		Kubernetes:       kubernetes,
		ContainerRuntime: containerRuntime,
	}

	// CommonParams
	log.Info("Creating common params")
	cparams := chaos.CommonParams{
		Timeout:       60,
		Namespace:     namespace,
		LabelSelector: labelSelector,
	}

	// NetworkLossParams
	log.Info("Creating network-loss params")
	fparams := chaos.NetworkLossParams{
		DestinationIPs:   destinationIPs,
		NetworkInterface: networkInterface,
		NetemCommands:    netemCommands,
	}

	networkLossFaulEngine.Execute(ctx, cparams, fparams)

	time.Sleep(600 * time.Second)
}

func getClientset() *kubernetes.Clientset {
	log.Info("Getting InCluster config for clientset")
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("getClientset - error on getting InCluster config for clientset: %s", err)
	}

	log.Info("Creating new clientset")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClientset - error on creating new clientset: %s", err)
	}

	return clientset
}

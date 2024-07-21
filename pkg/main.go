package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/chaos"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/client"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	ctx := context.Background()

	fmt.Println("sinth-chaos-poc")

	namespace := os.Getenv("NAMESPACE")
	labelSelector := os.Getenv("LABEL_SELECTOR")
	networkInterface := os.Getenv("NETWORK_INTERFACE")
	netemCommands := os.Getenv("NETEM_COMMANDS")
	socketPath := os.Getenv("SOCKET_PATH")
	destinationIPs := os.Getenv("DESTINATION_IPS")

	// KubernetesClient
	var kubernetesClient client.KubernetesClient = client.NewK8sClient(getClientset())

	// Command
	var command cmd.Command = cmd.NewOsBashExec()

	// ContainerRuntimeClient
	var containerRuntimeClient client.ContainerRuntimeClient = client.NewContainerdClient(command, socketPath)

	// NetworkLossParams
	params := chaos.NetworkLossParams{
		Namespace:        namespace,
		LabelSelector:    labelSelector,
		DestinationIPs:   destinationIPs,
		NetworkInterface: networkInterface,
		NetemCommands:    netemCommands,
	}

	// NetworkLoss
	networkLoss := chaos.NewNetworkLoss(kubernetesClient, containerRuntimeClient, command)
	networkLoss.Execute(ctx, params)

	time.Sleep(600 * time.Second)
}

func getClientset() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		fmt.Println("error on getClientset")
		panic(err)
	}

	return clientset
}

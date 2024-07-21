package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/client"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type InfoDetails struct {
	PID int `json:"pid"`
}

type CrictlInspectResponse struct {
	Info InfoDetails `json:"info"`
}

func main() {
	ctx := context.Background()

	fmt.Println("sinth-chaos-poc")

	namespace := os.Getenv("NAMESPACE")
	labelSelector := os.Getenv("LABEL_SELECTOR")
	netInterface := os.Getenv("NETWORK_INTERFACE")
	netemCommands := os.Getenv("NETEM_COMMANDS")
	socketPath := os.Getenv("SOCKET_PATH")
	destinationIPs := os.Getenv("DESTINATION_IPS")

	// K8sClient
	clientset, err := getClientset()
	if err != nil {
		fmt.Println("error on getClientset")
		panic(err)
	}
	k8sClient := client.NewK8sClient(clientset)

	// OsBashExec
	osBashExec := cmd.NewOsBashExec()

	// Containerd client
	containerdClient := client.NewContainerdClient(osBashExec, socketPath)

	// Get Pods
	pods, err := k8sClient.GetPodsByNamespaceAndLabelSelector(ctx, namespace, labelSelector)
	if err != nil {
		fmt.Println("error on GetPodsByNamespaceAndLabelSelector")
		panic(err)
	}

	// Get first container of each pod
	containerIds := getPodsContainerIds(pods)

	// Get pid of each container
	pids, err := getCrioPIDs(ctx, containerdClient, containerIds)
	if err != nil {
		fmt.Printf("error on getCrioPIDs: %s\n", err.Error())
		panic(err)
	}

	fmt.Printf("injecting pids %v...\n", pids)

	for _, pid := range pids {
		var injects []string
		destinationIPsSlice := strings.Split(destinationIPs, ",")

		if len(destinationIPsSlice) == 0 {
			inject := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s root netem %v", pid, netInterface, netemCommands)
			injects = append(injects, inject)
		} else {
			inject := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev eth0 root handle 1: prio", pid)
			injects = append(injects, inject)

			inject = fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev eth0 parent 1:3 netem loss 100", pid)
			injects = append(injects, inject)

			for _, destinationIP := range destinationIPsSlice {
				inject = fmt.Sprintf("sudo nsenter -t %d -n tc filter add dev eth0 protocol ip parent 1:0 prio 3 u32 match ip dst %s flowid 1:3", pid, destinationIP)
				injects = append(injects, inject)
			}
		}

		fmt.Printf("injecting on pid %d\n", pid)

		for _, inject := range injects {
			_, _, err := osBashExec.Exec(inject)
			if err != nil {
				fmt.Println("error on inject cmd")
				return
			}
		}
	}

	time.Sleep(60 * time.Second)

	fmt.Printf("killing pids %v...\n", pids)

	for _, pid := range pids {
		kill := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc delete dev %s root", pid, netInterface)

		fmt.Printf("killing on pid %d\n", pid)

		_, _, err := osBashExec.Exec(kill)
		if err != nil {
			fmt.Println("error on kill cmd")
			return
		}
	}

	fmt.Println("end")

	time.Sleep(600 * time.Second)
}

func getClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	return kubernetes.NewForConfig(config)
}

func getPodsContainerIds(pods []client.Pod) []string {
	var containerIds []string
	for _, pod := range pods {
		firstContainer := pod.Containers[0]
		containerIds = append(containerIds, firstContainer.ID)
	}

	return containerIds
}

func getCrioPIDs(ctx context.Context, containerdClient client.ContainerdClient, containerIDs []string) ([]int, error) {
	var pids []int
	for _, containerID := range containerIDs {
		container, err := containerdClient.GetContainerByID(ctx, containerID)
		if err != nil {
			return nil, err
		}
		pids = append(pids, container.PID)
	}
	return pids, nil
}

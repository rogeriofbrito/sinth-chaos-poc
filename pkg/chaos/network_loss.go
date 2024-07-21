package chaos

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/client"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
)

type NetworkLossParams struct {
	Namespace        string
	LabelSelector    string
	DestinationIPs   string
	NetworkInterface string
	NetemCommands    string
}

type NetworkLoss struct {
	KubernetesClient       client.KubernetesClient
	ContainerRuntimeClient client.ContainerRuntimeClient
	Command                cmd.Command
}

func NewNetworkLoss(kubernetesClient client.KubernetesClient, containerRuntimeClient client.ContainerRuntimeClient, command cmd.Command) NetworkLoss {
	return NetworkLoss{
		KubernetesClient:       kubernetesClient,
		ContainerRuntimeClient: containerRuntimeClient,
		Command:                command,
	}
}

func (networkLoss NetworkLoss) Execute(ctx context.Context, params NetworkLossParams) {
	ctx, cancelFunc := context.WithTimeout(ctx, time.Duration(60*time.Second))
	defer cancelFunc()

	wg := sync.WaitGroup{}

	var injectedPods = make(map[string]bool)

Loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("end chaos")
			break Loop
		default:
			fmt.Println("searching pods")

			pods, err := networkLoss.KubernetesClient.GetPodsByNamespaceAndLabelSelector(ctx, params.Namespace, params.LabelSelector)
			if err != nil {
				cancelFunc()
			}

			for _, pod := range pods {
				if !injectedPods[pod.Name] {
					wg.Add(1)
					go networkLoss.do(ctx, &wg, pod, params)
					injectedPods[pod.Name] = true
				}
			}

			time.Sleep(1 * time.Second)
		}
	}

	wg.Wait()
}

func (networkLoss NetworkLoss) do(ctx context.Context, wg *sync.WaitGroup, pod client.Pod, params NetworkLossParams) {
	defer wg.Done()

	// inject

	containerID := pod.Containers[0].ID

	container, err := networkLoss.ContainerRuntimeClient.GetContainerByID(ctx, containerID)
	if err != nil {
		return
	}

	pid := container.PID

	err = networkLoss.inject(pid, params)
	if err != nil {
		return
	}

	<-ctx.Done()

	// kill

	err = networkLoss.kill(pid, params)
	if err != nil {
		return
	}
}

func (networkLoss NetworkLoss) inject(pid int, params NetworkLossParams) error {
	var injects []string
	destinationIPsSlice := strings.Split(params.DestinationIPs, ",")

	if len(destinationIPsSlice) == 0 {
		inject := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s root netem %s", pid, params.NetworkInterface, params.NetemCommands)
		injects = append(injects, inject)
	} else {
		inject := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s root handle 1: prio", pid, params.NetworkInterface)
		injects = append(injects, inject)

		inject = fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s parent 1:3 netem %s", pid, params.NetworkInterface, params.NetemCommands)
		injects = append(injects, inject)

		for _, destinationIP := range destinationIPsSlice {
			inject = fmt.Sprintf("sudo nsenter -t %d -n tc filter add dev %s protocol ip parent 1:0 prio 3 u32 match ip dst %s flowid 1:3", pid, params.NetworkInterface, destinationIP)
			injects = append(injects, inject)
		}
	}

	fmt.Printf("injecting on pid %d\n", pid)

	for _, inject := range injects {
		_, _, err := networkLoss.Command.Exec(inject)
		if err != nil {
			return err
		}
	}

	return nil
}

func (networkLoss NetworkLoss) kill(pid int, params NetworkLossParams) error {
	kill := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc delete dev %s root", pid, params.NetworkInterface)

	fmt.Printf("killing on pid %d\n", pid)

	_, _, err := networkLoss.Command.Exec(kill)
	if err != nil {
		return err
	}

	return nil
}

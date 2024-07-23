package chaos

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/client"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/log"
)

type NetworkLossParams struct {
	Namespace        string
	LabelSelector    string
	DestinationIPs   string
	NetworkInterface string
	NetemCommands    string
}

func (networkLossParams NetworkLossParams) String() string {
	bytes, _ := json.Marshal(networkLossParams)
	return string(bytes)
}

type NetworkLoss struct {
	Kubernetes       client.Kubernetes
	ContainerRuntime client.ContainerRuntime
	Command          cmd.Command
}

func NewNetworkLoss(Kubernetes client.Kubernetes, ContainerRuntime client.ContainerRuntime, command cmd.Command) NetworkLoss {
	log.Info("Creating new NetworkLoss chaos ")
	return NetworkLoss{
		Kubernetes:       Kubernetes,
		ContainerRuntime: ContainerRuntime,
		Command:          command,
	}
}

func (n NetworkLoss) Execute(ctx context.Context, params NetworkLossParams) {
	log.Infof("Executing network-loss chaos with params: %s", params.String())

	ctx, cancelFunc := context.WithTimeout(ctx, time.Duration(60*time.Second))
	defer cancelFunc()

	wg := sync.WaitGroup{}

	var isProcessed = make(map[string]bool)

Loop:
	for {
		select {
		case <-ctx.Done():
			log.Info("Chaos end")
			break Loop
		default:
			pods := n.searchPods(ctx, cancelFunc, params)
			n.injectAndRemoveFaultEveryPod(ctx, cancelFunc, &wg, &isProcessed, pods, params)
			time.Sleep(1 * time.Second)
		}
	}

	wg.Wait()
}

func (n NetworkLoss) searchPods(ctx context.Context, cancelFunc context.CancelFunc, params NetworkLossParams) []client.Pod {
	pods, err := n.Kubernetes.GetPodsByNamespaceAndLabelSelector(ctx, params.Namespace, params.LabelSelector)
	if err != nil {
		log.Errorf("NewNetworkLoss.searchPods - error on get pods by namespace and label selector: %s", err)
		cancelFunc()
	}
	return pods
}

func (n NetworkLoss) injectAndRemoveFaultEveryPod(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup, isProcessed *map[string]bool, pods []client.Pod, params NetworkLossParams) {
	for _, pod := range pods {
		if (*isProcessed)[pod.Name] {
			log.Infof("Fault already injected in pod %s", pod.Name)
			continue
		}

		go n.injectAndRemoveFault(ctx, cancelFunc, wg, pod, params)

		wg.Add(1)
		(*isProcessed)[pod.Name] = true
	}
}

func (n NetworkLoss) injectAndRemoveFault(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup, pod client.Pod, params NetworkLossParams) {
	defer wg.Done()

	containerName := pod.Containers[0].ID
	containerID := pod.Containers[0].ID

	log.Infof("Injecting and removing network-loss fault in pod %s - container name %s - container ID %s", pod.Name, containerName, containerID)

	container, err := n.ContainerRuntime.GetContainerByID(ctx, containerID)
	if err != nil {
		log.Errorf("NetworkLoss.injectAndRemoveFault - error on get container by ID %s", err)
		cancelFunc()
		return
	}

	log.Info("Injecting network-loss")

	err = n.injectFault(container.PID, params)
	if err != nil {
		log.Errorf("NetworkLoss.injectAndRemoveFault - error on inject fault on PID %d, error: %s", container.PID, err)
		cancelFunc()
		return
	}

	<-ctx.Done()

	log.Info("Removing network-loss")

	err = n.removeFault(container.PID, params)
	if err != nil {
		log.Errorf("NetworkLoss.injectAndRemoveFault - error on remove fault on PID %d, error: %s", container.PID, err)
		cancelFunc()
		return
	}
}

func (n NetworkLoss) injectFault(pid int, params NetworkLossParams) error {
	log.Infof("Injecting network-loss fault in pid %d", pid)

	var injects []string
	destinationIPsSlice := strings.Split(params.DestinationIPs, ",")
	log.Infof("Number of IPs: %d", len(destinationIPsSlice))

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

	for _, inject := range injects {
		_, _, err := n.Command.Exec(inject)
		if err != nil {
			log.Errorf("NetworkLoss.injectFault - error on exec inject fault command: %s", err)
			return err
		}
	}

	return nil
}

func (n NetworkLoss) removeFault(pid int, params NetworkLossParams) error {
	log.Infof("Removing network-loss fault in pid %d", pid)

	kill := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc delete dev %s root", pid, params.NetworkInterface)

	_, _, err := n.Command.Exec(kill)
	if err != nil {
		log.Errorf("NetworkLoss.removeFault - error on exec remove fault command: %s", err)
		return err
	}

	return nil
}

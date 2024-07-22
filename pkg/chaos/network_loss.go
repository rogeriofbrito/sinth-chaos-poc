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

	var injectedPods = make(map[string]bool)

Loop:
	for {
		select {
		case <-ctx.Done():
			log.Info("Context done, chaos end")
			break Loop
		default:
			pods, err := n.Kubernetes.GetPodsByNamespaceAndLabelSelector(ctx, params.Namespace, params.LabelSelector)
			if err != nil {
				log.Errorf("NewNetworkLoss.Execute - error on get pods by namespace and label selector: %s", err)
				cancelFunc()
			}

			for _, pod := range pods {
				if !injectedPods[pod.Name] {
					wg.Add(1)
					go n.do(ctx, cancelFunc, &wg, pod, params)
					injectedPods[pod.Name] = true
				} else {
					log.Infof("Fault already injected in pod %s", pod.Name)
				}
			}

			time.Sleep(1 * time.Second)
		}
	}

	wg.Wait()
}

func (n NetworkLoss) do(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup, pod client.Pod, params NetworkLossParams) {
	defer wg.Done()

	log.Infof("Injecting network-loss chaos in pod %s", pod.Name)

	containerID := pod.Containers[0].ID

	log.Infof("Container id %s", containerID)

	container, err := n.ContainerRuntime.GetContainerByID(ctx, containerID)
	if err != nil {
		log.Errorf("NetworkLoss.do - error on get container by ID: %s", err)
		cancelFunc()
		return
	}

	pid := container.PID
	log.Infof("Container PID: %d", pid)

	err = n.inject(pid, params)
	if err != nil {
		log.Errorf("NetworkLoss.do - error on inject chaos on PID %d, error: %s", pid, err)
		cancelFunc()
		return
	}

	<-ctx.Done()

	log.Infof("Killling network-loss chaos in pod %s", pod.Name)

	err = n.kill(pid, params)
	if err != nil {
		log.Errorf("NetworkLoss.do - error on kill chaos on PID %d, error: %s", pid, err)
		cancelFunc()
		return
	}
}

func (n NetworkLoss) inject(pid int, params NetworkLossParams) error {
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
			log.Errorf("NetworkLoss.inject - error on exec inject command: %s", err)
			return err
		}
	}

	return nil
}

func (n NetworkLoss) kill(pid int, params NetworkLossParams) error {
	kill := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc delete dev %s root", pid, params.NetworkInterface)

	_, _, err := n.Command.Exec(kill)
	if err != nil {
		log.Errorf("NetworkLoss.kill - error on exec inject command: %s", err)
		return err
	}

	return nil
}

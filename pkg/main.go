package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	k8sClient, err := getK8sClient()
	if err != nil {
		fmt.Println("error on getK8sClient")
		panic(err)
	}

	pods, err := getPods(ctx, k8sClient, namespace, labelSelector)
	if err != nil {
		fmt.Println("error on getK8sClient")
		panic(err)
	}

	containerIds := getPodsContainerIds(pods)
	pids, err := getCrioPIDs(containerIds, socketPath)
	if err != nil {
		fmt.Printf("error on getCrioPIDs: %s\n", err.Error())
		panic(err)
	}

	fmt.Printf("injecting pids %v...\n", pids)

	for _, pid := range pids {
		var injectCmds []*exec.Cmd
		destinationIPsSlice := strings.Split(destinationIPs, ",")

		if len(destinationIPsSlice) == 0 {
			inject := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s root netem %v", pid, netInterface, netemCommands)
			injectCmds = append(injectCmds, exec.Command("/bin/bash", "-c", inject))
		} else {
			inject := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev eth0 root handle 1: prio", pid)
			injectCmds = append(injectCmds, exec.Command("/bin/bash", "-c", inject))

			inject = fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev eth0 parent 1:3 netem loss 100", pid)
			injectCmds = append(injectCmds, exec.Command("/bin/bash", "-c", inject))

			for _, destinationIP := range destinationIPsSlice {
				inject = fmt.Sprintf("sudo nsenter -t %d -n tc filter add dev eth0 protocol ip parent 1:0 prio 3 u32 match ip dst %s flowid 1:3", pid, destinationIP)
				injectCmds = append(injectCmds, exec.Command("/bin/bash", "-c", inject))
			}
		}

		fmt.Printf("injecting on pid %d\n", pid)

		for _, injectCmd := range injectCmds {
			_, _, err := runCmd(injectCmd)
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
		killCmd := exec.Command("/bin/bash", "-c", kill)

		fmt.Printf("killing on pid %d\n", pid)

		_, _, err := runCmd(killCmd)
		if err != nil {
			fmt.Println("error on kill cmd")
			return
		}
	}

	fmt.Println("end")

	time.Sleep(600 * time.Second)
}

func runCmd(cmd *exec.Cmd) (*bytes.Buffer, *bytes.Buffer, error) {
	var out, stdErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		return nil, nil, err
	}

	return &out, &stdErr, nil
}

func getK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	return kubernetes.NewForConfig(config)
}

func getPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string, labelSelector string) (*v1.PodList, error) {
	return clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
}

func getPodsContainerIds(pods *v1.PodList) []string {
	var containerIds []string
	for _, pod := range pods.Items {
		firstContainer := pod.Status.ContainerStatuses[0]
		containerID := strings.Replace(firstContainer.ContainerID, "containerd://", "", 1) // remove containerd:// prefix
		containerIds = append(containerIds, containerID)
	}

	return containerIds
}

func getCrioPIDs(containerIDs []string, socketPath string) ([]int, error) {
	var pids []int
	for _, containerID := range containerIDs {
		pid, err := getCrioPID(containerID, socketPath)
		if err != nil {
			return nil, err
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func getCrioPID(containerID string, socketPath string) (int, error) {
	var pid int
	cmd := exec.Command("sudo", "crictl", "-i", fmt.Sprintf("unix://%s", socketPath), "-r", fmt.Sprintf("unix://%s", socketPath), "inspect", containerID)
	fmt.Printf("cmd: %s\n", cmd.String())
	out, err := inspect(cmd)
	if err != nil {
		fmt.Printf("error on crictl run: %s\n", err.Error())
		return 0, err
	}

	var resp CrictlInspectResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		fmt.Printf("error on unmarshal crictl response: %s\n", err.Error())
		return 0, err
	}
	pid = resp.Info.PID

	return pid, nil
}

func inspect(cmd *exec.Cmd) ([]byte, error) {
	var out, stdErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stdErr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

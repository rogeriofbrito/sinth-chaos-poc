package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	fmt.Println("sinth-chaos-poc")

	pids := os.Getenv("PIDS")
	netInterface := os.Getenv("NETWORK_INTERFACE")
	netemCommands := os.Getenv("NETEM_COMMANDS")

	fmt.Println("injecting...")

	for _, pid := range strings.Split(pids, ",") {
		inject := fmt.Sprintf("sudo nsenter -t %s -n tc qdisc replace dev %s root netem %v", pid, netInterface, netemCommands)
		injectCmd := exec.Command("/bin/bash", "-c", inject)

		fmt.Printf("injecting on pid %s\n", pid)

		_, _, err := runCmd(injectCmd)
		if err != nil {
			fmt.Println("error on inject cmd")
			return
		}
	}

	time.Sleep(60 * time.Second)

	fmt.Println("killing...")

	for _, pid := range strings.Split(pids, ",") {
		kill := fmt.Sprintf("sudo nsenter -t %s -n tc qdisc delete dev %s root", pid, netInterface)
		killCmd := exec.Command("/bin/bash", "-c", kill)

		fmt.Printf("killing on pid %s\n", pid)

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

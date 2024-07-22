package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/log"
)

type InfoDetails struct { // TODO: change to private
	PID int `json:"pid"`
}

type CrictlInspectResponse struct { // TODO: change to private
	Info InfoDetails `json:"info"`
}

type ContainerdClient struct {
	command    cmd.Command
	socketPath string
}

func NewContainerdClient(command cmd.Command, socketPath string) ContainerdClient {
	log.Info("Creating new container runtime Client (ContainerdClient)")
	return ContainerdClient{
		command:    command,
		socketPath: socketPath,
	}
}

func (containerdClient ContainerdClient) GetContainerByID(ctx context.Context, containerID string) (Container, error) {
	log.Infof("Getting container by ID (%s)", containerID)

	cmd := fmt.Sprintf("sudo crictl -i unix://%s -r unix://%s inspect %s", containerdClient.socketPath, containerdClient.socketPath, containerID)

	log.Info(cmd)

	stdout, stderr, err := containerdClient.command.Exec(cmd)
	if err != nil {
		log.Errorf("Error on executing crictl inspect command: %s, stderr: %s", err, stderr)
		return Container{}, err
	}
	log.Infof("Output crictl inspect: %s", stdout)

	var resp CrictlInspectResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		log.Infof("Error on unmarshal crictl inspect response: %s", err)
		return Container{}, err
	}

	container := Container{
		PID: resp.Info.PID,
	}

	return container, nil
}

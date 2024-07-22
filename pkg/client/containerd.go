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

func (containerdClient ContainerdClient) GetContainerByID(ctx context.Context, containerID string) (Container, error) { // TODO: change to pointer
	log.Infof("Getting container by ID (%s)", containerID)

	cmd := fmt.Sprintf("sudo crictl -i unix://%s -r unix://%s inspect %s", containerdClient.socketPath, containerdClient.socketPath, containerID)

	log.Info(cmd)

	stdout, stderr, err := containerdClient.command.Exec(cmd)
	if err != nil {
		return Container{}, fmt.Errorf("ContainerdClient.GetContainerByID - error on executing crictl inspect command: %w, stderr: %s", err, stderr)
	}
	log.Infof("Output crictl inspect: %s", stdout)

	var resp CrictlInspectResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		return Container{}, fmt.Errorf("ContainerdClient.GetContainerByID - error on unmarshal crictl inspect response: %w", err)
	}

	container := Container{
		PID: resp.Info.PID,
	}

	return container, nil
}

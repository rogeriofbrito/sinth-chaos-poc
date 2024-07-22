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

type ContainerdContainerRuntime struct {
	command    cmd.Command
	socketPath string
}

func NewContainerdContainerRuntime(command cmd.Command, socketPath string) ContainerdContainerRuntime {
	log.Info("Creating new ContainerdContainerRuntime client (ContainerdContainerRuntime)")
	return ContainerdContainerRuntime{
		command:    command,
		socketPath: socketPath,
	}
}

func (c ContainerdContainerRuntime) GetContainerByID(ctx context.Context, containerID string) (Container, error) { // TODO: change to pointer
	log.Infof("Getting container by ID (%s)", containerID)

	cmd := fmt.Sprintf("sudo crictl -i unix://%s -r unix://%s inspect %s", c.socketPath, c.socketPath, containerID)

	log.Info(cmd)

	stdout, stderr, err := c.command.Exec(cmd)
	if err != nil {
		return Container{}, fmt.Errorf("ContainerdContainerRuntime.GetContainerByID - error on executing crictl inspect command: %w, stderr: %s", err, stderr)
	}
	log.Infof("Output crictl inspect: %s", stdout)

	var resp CrictlInspectResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		return Container{}, fmt.Errorf("ContainerdContainerRuntime.GetContainerByID - error on unmarshal crictl inspect response: %w", err)
	}

	container := Container{
		PID: resp.Info.PID,
	}

	return container, nil
}

package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
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
	return ContainerdClient{
		command:    command,
		socketPath: socketPath,
	}
}

func (containerdClient ContainerdClient) GetContainerByID(ctx context.Context, containerID string) (Container, error) {
	cmd := fmt.Sprintf("sudo crictl -i unix://%s -r unix://%s inspect %s", containerdClient.socketPath, containerdClient.socketPath, containerID)

	stdout, _, err := containerdClient.command.Exec(cmd)
	if err != nil {
		return Container{}, err
	}

	var resp CrictlInspectResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		return Container{}, err
	}

	container := Container{
		PID: resp.Info.PID,
	}

	return container, nil
}

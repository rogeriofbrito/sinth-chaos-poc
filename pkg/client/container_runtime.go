package client

import "context"

type ContainerRuntime interface {
	GetContainerByID(ctx context.Context, containerID string) (Container, error)
}

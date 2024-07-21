package client

import "context"

type ContainerRuntimeClient interface {
	GetContainerByID(ctx context.Context, containerID string) (Container, error)
}

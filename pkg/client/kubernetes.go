package client

import (
	"context"
)

type Kubernetes interface {
	GetPodsByNamespaceAndLabelSelector(ctx context.Context, namespace string, labelSelector string) ([]Pod, error)
}

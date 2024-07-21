package client

import (
	"context"
)

type KubernetesClient interface {
	GetPodsByNamespaceAndLabelSelector(ctx context.Context, namespace string, labelSelector string) ([]Pod, error)
}

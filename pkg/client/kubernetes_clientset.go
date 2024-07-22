package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClientsetKubernetes struct {
	clientset *kubernetes.Clientset
}

func NewClientsetKubernetes(clientset *kubernetes.Clientset) *ClientsetKubernetes {
	log.Info("Creating new ClientsetKubernetes Client")
	return &ClientsetKubernetes{
		clientset: clientset,
	}
}

func (c ClientsetKubernetes) GetPodsByNamespaceAndLabelSelector(ctx context.Context, namespace string, labelSelector string) ([]Pod, error) {
	log.Infof("Getting pods by namespace (%s) and label selector (%s)", namespace, labelSelector)
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("ClientsetKubernetes.GetPodsByNamespaceAndLabelSelector - error on listing pods: %w", err)
	}

	log.Infof("%d pods were found", len(podList.Items))

	var pods []Pod
	for _, podItem := range podList.Items {
		var podContainers []PodContainer
		// TODO: filter only running pods
		log.Infof("Getting containers of pod %s", podItem.Name)
		for _, containerStatuses := range podItem.Status.ContainerStatuses {
			podContainer := PodContainer{
				ID: c.getContainerIDFromContainerdUri(containerStatuses.ContainerID),
			}
			podContainers = append(podContainers, podContainer)
		}

		pod := Pod{
			Name:       podItem.Name,
			Containers: podContainers,
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

func (c ClientsetKubernetes) getContainerIDFromContainerdUri(containerdUri string) string {
	return strings.Replace(containerdUri, "containerd://", "", 1) // remove containerd:// prefix
}

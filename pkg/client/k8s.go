package client

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sClient struct {
	clientset *kubernetes.Clientset
}

func NewK8sClient(clientset *kubernetes.Clientset) *K8sClient {
	return &K8sClient{
		clientset: clientset,
	}
}

func (k8sClient K8sClient) GetPodsByNamespaceAndLabelSelector(ctx context.Context, namespace string, labelSelector string) ([]Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	podList, err := k8sClient.clientset.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	var pods []Pod
	for _, podItem := range podList.Items {
		var podContainers []PodContainer
		for _, containerStatuses := range podItem.Status.ContainerStatuses {
			podContainer := PodContainer{
				ID: k8sClient.getContainerIDFromContainerdUri(containerStatuses.ContainerID),
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

func (k8sClient K8sClient) getContainerIDFromContainerdUri(containerdUri string) string {
	return strings.Replace(containerdUri, "containerd://", "", 1) // remove containerd:// prefix
}

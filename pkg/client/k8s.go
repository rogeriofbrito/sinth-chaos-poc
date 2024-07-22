package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sClient struct {
	clientset *kubernetes.Clientset
}

func NewK8sClient(clientset *kubernetes.Clientset) *K8sClient {
	log.Info("Creating new Kubernetes Client (K8sClient)")
	return &K8sClient{
		clientset: clientset,
	}
}

func (k8sClient K8sClient) GetPodsByNamespaceAndLabelSelector(ctx context.Context, namespace string, labelSelector string) ([]Pod, error) {
	log.Infof("Getting pods by namespace (%s) and label selector (%s)", namespace, labelSelector)
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	podList, err := k8sClient.clientset.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("K8sClient.GetPodsByNamespaceAndLabelSelector - error on listing pods: %w", err)
	}

	log.Infof("%d pods were found", len(podList.Items))

	var pods []Pod
	for _, podItem := range podList.Items {
		var podContainers []PodContainer
		// TODO: filter only running pods
		log.Info("Getting containers from pods")
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

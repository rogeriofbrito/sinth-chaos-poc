package chaos

import (
	"context"
	"sync"
	"time"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/client"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/log"
)

type FaultEngine struct {
	FaultAction      IFaultAction
	Kubernetes       client.Kubernetes
	ContainerRuntime client.ContainerRuntime
}

func NewFaultEngine(faultAction IFaultAction, Kubernetes client.Kubernetes, containerRuntime client.ContainerRuntime) FaultEngine {
	return FaultEngine{
		FaultAction:      faultAction,
		Kubernetes:       Kubernetes,
		ContainerRuntime: containerRuntime,
	}
}

func (f FaultEngine) Execute(ctx context.Context, cparams CommonParams, fparams interface{}) {
	log.Infof("Executing FaultEngine with cparams (%v) and fparams (%v)", cparams, fparams)

	ctx, cancelFunc := context.WithTimeout(ctx, time.Duration(cparams.Timeout*int64(time.Second)))
	defer cancelFunc()

	wg := sync.WaitGroup{}

	var isProcessed = make(map[string]bool)

Loop:
	for {
		select {
		case <-ctx.Done():
			log.Info("FaultEngine end execution")
			break Loop
		default:
			pods := f.getPods(ctx, cancelFunc, cparams)
			if pods != nil {
				f.injectAndRemoveFaultEveryPod(ctx, cancelFunc, &wg, &isProcessed, pods, fparams)
				time.Sleep(1 * time.Second)
			}
		}
	}

	wg.Wait()
}

func (f FaultEngine) getPods(ctx context.Context, cancelFunc context.CancelFunc, cparams CommonParams) []client.Pod {
	pods, err := f.Kubernetes.GetPodsByNamespaceAndLabelSelector(ctx, cparams.Namespace, cparams.LabelSelector)
	if err != nil {
		log.Errorf("FaultEngine.searchPods - error on get pods by namespace and label selector: %s", err)
		cancelFunc()
		return nil
	}
	return pods
}

func (f FaultEngine) getContainer(ctx context.Context, cancelFunc context.CancelFunc, containerID string) *client.Container {
	container, err := f.ContainerRuntime.GetContainerByID(ctx, containerID)
	if err != nil {
		log.Errorf("FaultEngine.getContainer - error on get container by ID %s", err)
		cancelFunc()
		return nil
	}
	return &container
}

func (f FaultEngine) injectAndRemoveFaultEveryPod(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup, isProcessed *map[string]bool, pods []client.Pod, fparams interface{}) {
	for _, pod := range pods {
		if (*isProcessed)[pod.Name] {
			log.Infof("Fault already injected in pod %s", pod.Name)
			continue
		}

		go f.injectAndRemoveFault(ctx, cancelFunc, wg, pod, fparams)

		wg.Add(1)
		(*isProcessed)[pod.Name] = true
	}
}

func (f FaultEngine) injectAndRemoveFault(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup, pod client.Pod, fparams interface{}) {
	defer wg.Done()

	containerName := pod.Containers[0].ID // TODO: in some scenarios, pod may not have containers in execution
	containerID := pod.Containers[0].ID   // TODO: in some scenarios, pod may not have containers in execution

	log.Infof("Injecting and removing network-loss fault in pod (%s), containerName (%s) and containerID (%s)", pod.Name, containerName, containerID)

	container := f.getContainer(ctx, cancelFunc, containerID)

	if container != nil {
		f.injectFault(cancelFunc, container.PID, fparams)
		<-ctx.Done()
		f.removeFault(cancelFunc, container.PID, fparams)
	}
}

func (f FaultEngine) injectFault(cancelFunc context.CancelFunc, pid int, fparams interface{}) {
	log.Info("Injecting fault")

	err := f.FaultAction.Inject(pid, fparams)
	if err != nil {
		log.Errorf("FaultEngine.injectFault - error on inject fault on PID %d, error: %s", pid, err)
		cancelFunc()
		return
	}
}

func (f FaultEngine) removeFault(cancelFunc context.CancelFunc, pid int, fparams interface{}) {
	log.Info("Removing fault")

	err := f.FaultAction.Remove(pid, fparams)
	if err != nil {
		log.Errorf("FaultEngine.removeFault - error on remove fault on PID %d, error: %s", pid, err)
		cancelFunc()
		return
	}
}

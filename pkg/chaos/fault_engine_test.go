package chaos_test

import (
	"context"
	"testing"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/chaos"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_chaos "github.com/rogeriofbrito/sinth-chaos-poc/mocks/github.com/rogeriofbrito/sinth-chaos-poc/pkg/chaos"
	mock_client "github.com/rogeriofbrito/sinth-chaos-poc/mocks/github.com/rogeriofbrito/sinth-chaos-poc/pkg/client"
)

var faultActionMock *mock_chaos.MockIFaultAction
var kubernetsMock *mock_client.MockKubernetes
var containerRuntimeMock *mock_client.MockContainerRuntime

var faultEngine chaos.FaultEngine
var ctx context.Context

func faultEngineInit(t *testing.T) {
	faultActionMock = mock_chaos.NewMockIFaultAction(t)
	kubernetsMock = mock_client.NewMockKubernetes(t)
	containerRuntimeMock = mock_client.NewMockContainerRuntime(t)

	faultEngine = chaos.FaultEngine{
		FaultAction:      faultActionMock,
		Kubernetes:       kubernetsMock,
		ContainerRuntime: containerRuntimeMock,
	}
	ctx = context.Background()
}

func TestFaultEngineSuccess(t *testing.T) {
	faultEngineInit(t)

	faultActionMock.EXPECT().Inject(mock.Anything, mock.Anything).Return(nil)
	faultActionMock.EXPECT().Remove(mock.Anything, mock.Anything).Return(nil)

	pods := []client.Pod{
		{
			Name: "pod-1",
			Containers: []client.PodContainer{
				{
					ID: "5de8497f85dc417da9d3d6841ba8494b",
				},
			},
		},
	}
	kubernetsMock.EXPECT().GetPodsByNamespaceAndLabelSelector(mock.Anything, mock.Anything, mock.Anything).Return(pods, nil)

	pid := 2345
	container := client.Container{
		PID: pid,
	}
	containerRuntimeMock.EXPECT().GetContainerByID(mock.Anything, mock.Anything).Return(container, nil)

	cparams := chaos.CommonParams{
		Timeout:       3,
		Namespace:     "app",
		LabelSelector: "app=my-app",
	}
	fparams := chaos.NetworkLossParams{
		DestinationIPs:   "127.0.0.1",
		NetworkInterface: "eth0",
		NetemCommands:    "loss 100",
	}

	faultEngine.Execute(ctx, cparams, fparams)

	assert.Len(t, faultActionMock.Calls, 2)
	assert.Equal(t, "Inject", faultActionMock.Calls[0].Method)
	assert.Equal(t, pid, faultActionMock.Calls[0].Arguments[0])
	assert.Equal(t, fparams, faultActionMock.Calls[0].Arguments[1])
	assert.Equal(t, "Remove", faultActionMock.Calls[1].Method)
	assert.Equal(t, pid, faultActionMock.Calls[1].Arguments[0])
	assert.Equal(t, fparams, faultActionMock.Calls[1].Arguments[1])

	for _, call := range kubernetsMock.Calls {
		assert.Equal(t, "GetPodsByNamespaceAndLabelSelector", call.Method)
		assert.NotNil(t, kubernetsMock.Calls[0].Arguments[0])
		assert.Equal(t, cparams.Namespace, call.Arguments[1])
		assert.Equal(t, cparams.LabelSelector, call.Arguments[2])
	}
}

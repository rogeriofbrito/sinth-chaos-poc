package chaos_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mock_client "github.com/rogeriofbrito/sinth-chaos-poc/mocks/github.com/rogeriofbrito/sinth-chaos-poc/pkg/client"
	mock_cmd "github.com/rogeriofbrito/sinth-chaos-poc/mocks/github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/chaos"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Network Loss", func() {
	var kubernetsMock *mock_client.MockKubernetes
	var containerRuntimeMock *mock_client.MockContainerRuntime
	var commandMock *mock_cmd.MockCommand

	var networkLoss chaos.NetworkLoss
	var ctx context.Context

	t := GinkgoT()

	BeforeEach(func() {
		kubernetsMock = mock_client.NewMockKubernetes(t)
		containerRuntimeMock = mock_client.NewMockContainerRuntime(t)
		commandMock = mock_cmd.NewMockCommand(t)

		networkLoss = chaos.NetworkLoss{
			Kubernetes:       kubernetsMock,
			ContainerRuntime: containerRuntimeMock,
			Command:          commandMock,
		}

		ctx = context.Background()
	})

	Describe("Execute Network Loss fault", func() {
		Context("with fail in get pods", func() {
			It("should fail fault injection", func() {
				kubernetsMock.EXPECT().GetPodsByNamespaceAndLabelSelector(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))
				params := chaos.NetworkLossParams{
					Namespace:        "app",
					LabelSelector:    "app=my-app",
					DestinationIPs:   "127.0.0.1",
					NetworkInterface: "eth0",
					NetemCommands:    "loss 100",
				}
				networkLoss.Execute(ctx, params)

				Expect(len(kubernetsMock.Calls)).To(Equal(1))
				Expect(kubernetsMock.Calls[0].Method).To(Equal("GetPodsByNamespaceAndLabelSelector"))
				Expect(len(kubernetsMock.Calls[0].Arguments)).To(Equal(3))
				_, ok := kubernetsMock.Calls[0].Arguments.Get(0).(context.Context).Deadline()
				Expect(ok).To(BeTrue())
				Expect(kubernetsMock.Calls[0].Arguments.Get(1)).To(Equal(params.Namespace))
				Expect(kubernetsMock.Calls[0].Arguments.Get(2)).To(Equal(params.LabelSelector))

				Expect(len(containerRuntimeMock.Calls)).To(Equal(0))
				Expect(len(commandMock.Calls)).To(Equal(0))
			})
		})
	})
})

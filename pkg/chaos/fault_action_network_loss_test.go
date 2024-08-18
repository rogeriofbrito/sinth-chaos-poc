package chaos_test

import (
	"fmt"
	"testing"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/chaos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_cmd "github.com/rogeriofbrito/sinth-chaos-poc/mocks/github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
)

var commandMock *mock_cmd.MockCommand

var faultAction chaos.IFaultAction

func faultActionNetworkLossInit(t *testing.T) {
	commandMock = mock_cmd.NewMockCommand(t)
	faultAction = chaos.NewNetworkLossFaultAction(commandMock)
}

func TestFaultActionInjectNetworkLossAllIpsSuccess(t *testing.T) {
	faultActionNetworkLossInit(t)

	commandMock.EXPECT().Exec(mock.Anything).Return("", "", nil)

	pid := 123
	fparams := chaos.NetworkLossParams{
		NetworkInterface: "eth0",
		NetemCommands:    "loss 100",
	}
	err := faultAction.Inject(pid, fparams)

	assert.Nil(t, err)
	assert.Len(t, commandMock.Calls, 3)
	assert.Equal(t, "Exec", commandMock.Calls[0].Method)
	assert.Equal(t, "Exec", commandMock.Calls[1].Method)
	assert.Equal(t, "Exec", commandMock.Calls[2].Method)
	assert.Equal(t, fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s root handle 1: prio", pid, fparams.NetworkInterface), commandMock.Calls[0].Arguments[0])
	assert.Equal(t, fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s parent 1:3 netem %s", pid, fparams.NetworkInterface, fparams.NetemCommands), commandMock.Calls[1].Arguments[0])
	assert.Equal(t, fmt.Sprintf("sudo nsenter -t %d -n tc filter add dev %s protocol ip parent 1:0 prio 3 u32 match ip dst  flowid 1:3", pid, fparams.NetworkInterface), commandMock.Calls[2].Arguments[0])
}

func TestFaultActionInjectNetworkLossSelectedIpsSuccess(t *testing.T) {
	faultActionNetworkLossInit(t)

	commandMock.EXPECT().Exec(mock.Anything).Return("", "", nil)

	pid := 123
	ip1 := "172.217.28.238"
	ip2 := "72.21.206.80"
	fparams := chaos.NetworkLossParams{
		NetworkInterface: "eth0",
		NetemCommands:    "loss 100",
		DestinationIPs:   fmt.Sprintf("%s,%s", ip1, ip2),
	}
	err := faultAction.Inject(pid, fparams)

	assert.Nil(t, err)
	assert.Len(t, commandMock.Calls, 4)
	assert.Equal(t, "Exec", commandMock.Calls[0].Method)
	assert.Equal(t, "Exec", commandMock.Calls[1].Method)
	assert.Equal(t, "Exec", commandMock.Calls[2].Method)
	assert.Equal(t, "Exec", commandMock.Calls[3].Method)
	assert.Equal(t, fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s root handle 1: prio", pid, fparams.NetworkInterface), commandMock.Calls[0].Arguments[0])
	assert.Equal(t, fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s parent 1:3 netem %s", pid, fparams.NetworkInterface, fparams.NetemCommands), commandMock.Calls[1].Arguments[0])
	assert.Equal(t, fmt.Sprintf("sudo nsenter -t %d -n tc filter add dev %s protocol ip parent 1:0 prio 3 u32 match ip dst %s flowid 1:3", pid, fparams.NetworkInterface, ip1), commandMock.Calls[2].Arguments[0])
	assert.Equal(t, fmt.Sprintf("sudo nsenter -t %d -n tc filter add dev %s protocol ip parent 1:0 prio 3 u32 match ip dst %s flowid 1:3", pid, fparams.NetworkInterface, ip2), commandMock.Calls[3].Arguments[0])
}

func TestFaultActionInjectNetworkLossCommandExecError(t *testing.T) {
	faultActionNetworkLossInit(t)

	execError := fmt.Errorf("error on exec")
	commandMock.EXPECT().Exec(mock.Anything).Return("", "", execError)

	pid := 123
	fparams := chaos.NetworkLossParams{
		NetworkInterface: "eth0",
		NetemCommands:    "loss 100",
	}
	err := faultAction.Inject(pid, fparams)

	assert.Equal(t, execError, err)
}

package chaos

import (
	"fmt"
	"strings"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/cmd"
	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/log"
)

type NetworkLossFaultAction struct {
	Command cmd.Command
}

func NewNetworkLossFaultAction(command cmd.Command) NetworkLossFaultAction {
	log.Info("Creating new NetworkLoss chaos ")
	return NetworkLossFaultAction{
		Command: command,
	}
}

func (n NetworkLossFaultAction) Inject(pid int, params interface{}) error {
	log.Infof("Injecting network-loss fault in pid %d", pid)

	nparams := params.(NetworkLossParams)

	var injects []string
	destinationIPsSlice := strings.Split(nparams.DestinationIPs, ",")
	log.Infof("Number of IPs: %d", len(destinationIPsSlice))

	if len(destinationIPsSlice) == 0 {
		inject := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s root netem %s", pid, nparams.NetworkInterface, nparams.NetemCommands)
		injects = append(injects, inject)
	} else {
		inject := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s root handle 1: prio", pid, nparams.NetworkInterface)
		injects = append(injects, inject)

		inject = fmt.Sprintf("sudo nsenter -t %d -n tc qdisc replace dev %s parent 1:3 netem %s", pid, nparams.NetworkInterface, nparams.NetemCommands)
		injects = append(injects, inject)

		for _, destinationIP := range destinationIPsSlice {
			inject = fmt.Sprintf("sudo nsenter -t %d -n tc filter add dev %s protocol ip parent 1:0 prio 3 u32 match ip dst %s flowid 1:3", pid, nparams.NetworkInterface, destinationIP)
			injects = append(injects, inject)
		}
	}

	for _, inject := range injects {
		_, _, err := n.Command.Exec(inject)
		if err != nil {
			log.Errorf("NetworkLoss.injectFault - error on exec inject fault command: %s", err)
			return err
		}
	}

	return nil
}

func (n NetworkLossFaultAction) Remove(pid int, params interface{}) error {
	log.Infof("Removing network-loss fault in pid %d", pid)

	nparams := params.(NetworkLossParams)

	kill := fmt.Sprintf("sudo nsenter -t %d -n tc qdisc delete dev %s root", pid, nparams.NetworkInterface)

	_, _, err := n.Command.Exec(kill)
	if err != nil {
		log.Errorf("NetworkLoss.removeFault - error on exec remove fault command: %s", err)
		return err
	}

	return nil
}

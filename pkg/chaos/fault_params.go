package chaos

type CommonParams struct {
	Timeout       int64
	Namespace     string
	LabelSelector string
}

type NetworkLossParams struct {
	DestinationIPs   string
	NetworkInterface string
	NetemCommands    string
}

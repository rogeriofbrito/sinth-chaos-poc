package client

type Container struct {
	PID int
}

type PodContainer struct {
	ID string
}

type Pod struct {
	Name       string
	Containers []PodContainer
}

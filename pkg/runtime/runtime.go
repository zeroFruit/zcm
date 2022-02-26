package runtime

type Status string

const (
	Creating Status = "creating"
	Created         = "created"
	Running         = "running"
	Stopped         = "stopped"
)

type State struct {
	OCIVersion  string
	Id          string
	Status      Status
	Pid         int
	Bundle      string
	Annotations map[string]string
}

package client

type InputConfig struct {
	Dest    string
	Path    string
	Payload string
}

func (i InputConfig) IsComplete() bool {
	return i.Dest != "" && i.Path != "" && i.Payload != ""
}

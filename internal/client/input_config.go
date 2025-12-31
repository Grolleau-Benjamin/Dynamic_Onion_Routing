package client

import "fmt"

type InputConfig struct {
	OnionPath string
	Dest      string
	Payload   string
}

func (ic InputConfig) IsComplete() bool {
	return ic.OnionPath != "" && ic.Dest != "" && ic.Payload != ""
}

func (ic InputConfig) String() string {
	return fmt.Sprintf("InputConfig{\n\tOnionPath: %s,\n\tDest: %s,\n\tPayload: %s\n}",
		ic.OnionPath,
		ic.Dest,
		ic.Payload,
	)
}

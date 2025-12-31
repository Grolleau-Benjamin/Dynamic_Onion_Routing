package tui

import "github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"

type State struct {
	Client *client.Client
	Config client.InputConfig
	Logs   []string
}

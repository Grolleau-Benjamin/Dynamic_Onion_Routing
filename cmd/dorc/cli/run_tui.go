package cli

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	tuiui "github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client_tui"
)

func runTUI(c *client.Client, input client.InputConfig) {
	_ = tuiui.Run(c, input)
}

package cli

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	tuiui "github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client_tui"
)

func runTUI(c *client.Client) {
	_ = tuiui.Run(c)
}

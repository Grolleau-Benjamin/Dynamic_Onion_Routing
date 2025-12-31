package cli

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/server"
	"github.com/spf13/cobra"
)

var (
	rootCommand = &cobra.Command{
		Use:   "dord",
		Short: "Dynamic Onion Routing daemon",
		Long: `Dynamic Onion Routing Daemon (dord)

This daemon implements the server side of the Dynamic Onion Routing protocol.
It listens for incoming packets and routes them across the network.
`,
		Run: Run,
	}
)

func Execute() error {
	return rootCommand.Execute()
}

func init() {}

func Run(cmd *cobra.Command, args []string) {
	server.Run()
}


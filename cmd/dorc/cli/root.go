package cli

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/spf13/cobra"
)

var (
	rootCommand = &cobra.Command{
		Use:   "dorc",
		Short: "Dynamic Onion Routing client",
		Long: `Dynamic Onion Routing Client (dorc)

This client implements the Dynamic Onion Routing protocol.
It can be used either in headless CLI mode or in interactive TUI mode.
`,
		Run: Run,
	}
)

func Execute() error {
	return rootCommand.Execute()
}

func init() {}

func Run(cmd *cobra.Command, args []string) {
	client.Run()
}

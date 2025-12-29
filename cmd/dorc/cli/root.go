package cli

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	"github.com/spf13/cobra"
)

var (
	tui      bool
	logLevel string

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

func init() {
	rootCommand.Flags().BoolVar(&tui, "tui", false, "Enable TUI mode")
	rootCommand.Flags().StringVar(&logLevel, "log-level", "info", "Set log level [debug, info, warn, error, off]")
}

func Run(cmd *cobra.Command, args []string) {
	c := client.New()

	if tui {
		logger.SetLogLevel(logger.Off)
		runTUI(c)
		return
	}

	logger.SetLogLevel(logger.ParseLevel(logLevel))
	runCLI(c)
}

package cli

import (
	"os"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	"github.com/spf13/cobra"
)

var (
	tui      bool
	logLevel string

	path    string
	dest    string
	payload string

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

	rootCommand.Flags().StringVar(
		&path,
		"onion-path",
		"",
		"Path of relays to use. e.g. [cafe::1]:62503,127.0.0.1:123 | 1.2.3.4:5678",
	)
	rootCommand.Flags().StringVar(
		&dest,
		"dest",
		"",
		"Final IP. e.g. 8.8.8.8:63",
	)

	rootCommand.Flags().StringVar(
		&payload,
		"payload",
		"",
		"Payload to send to dest (hex-encoded)",
	)
}

func Run(cmd *cobra.Command, args []string) {
	input := client.InputConfig{
		Path:    path,
		Dest:    dest,
		Payload: payload,
	}

	logger.SetLogLevel(logger.ParseLevel(logLevel))

	c := client.New()

	if tui {
		logger.SetLogLevel(logger.Off)
		runTUI(c, input)
		return
	}

	if input.IsComplete() {
		cmd.PrintErrln("Err: onion-path, dest, and payload are mandatory when --tui is disabled")
		os.Exit(1)
	}

	ctx, err := client.BuildSendContext(input.Dest, input.Path, []byte(input.Payload))
	if err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	runCLI(c, ctx)
}

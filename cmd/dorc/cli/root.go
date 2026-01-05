package cli

import (
	"os"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client/sinks/stdout"
	stui "github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client/sinks/tui"
	"github.com/spf13/cobra"
)

var (
	logLevel  string
	onionPath string
	dest      string
	payload   string

	tui bool

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

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCommand.Flags().StringVar(
		&logLevel,
		"log-level",
		"info",
		"Set log level [debug, info, warn, error, off]",
	)
	rootCommand.Flags().StringVar(&onionPath,
		"onion-path",
		"",
		"Path of relays to use. e.g. [cafe::1]:62503,127.0.0.1:123 | 1.2.3.4:5678",
	)
	rootCommand.Flags().StringVar(&dest,
		"dest",
		"",
		"Final IP. e.g. 8.8.8.8:63",
	)
	rootCommand.Flags().StringVar(&payload,
		"payload",
		"",
		"Payload to send to dest (hex-encoded)",
	)

	rootCommand.Flags().BoolVar(&tui,
		"tui",
		false,
		"Enable TUI mode",
	)
}

func Run(cmd *cobra.Command, args []string) {
	ic := client.InputConfig{
		OnionPath: onionPath,
		Dest:      dest,
		Payload:   payload,
	}

	c := client.New()

	type Sinker interface {
		Start() error
	}
	var s Sinker

	if tui {
		s = stui.New(c, ic)
	} else {
		if !ic.IsComplete() {
			cmd.PrintErrln("Err: some flags in required flags are missing (onion-path, dest, payload).")
			os.Exit(1)
		}
		s = stdout.New(c, ic)
	}

	if err := s.Start(); err != nil {
		cmd.PrintErrln("Error running client:", err)
		os.Exit(1)
	}
}

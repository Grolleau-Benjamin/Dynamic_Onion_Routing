package cli

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/server"
)

var (
	addr  string
	port  uint16
	idDir string

	logLevel string

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

func init() {
	rootCommand.Flags().StringVarP(
		&addr,
		"addr",
		"a",
		"::1",
		"IP address where the server will listen on",
	)

	rootCommand.Flags().Uint16VarP(
		&port,
		"port",
		"p",
		uint16(62503),
		"Port where the server will listen on",
	)

	rootCommand.Flags().StringVar(
		&idDir,
		"id-dir",
		"~/.dor",
		"Directory where identity material is stored",
	)

	rootCommand.Flags().StringVarP(
		&logLevel,
		"log-level",
		"l",
		"info",
		"Log level (debug, info, warn, error, off)",
	)
}

func Run(cmd *cobra.Command, args []string) {
	lvl := logger.ParseLevel(logLevel)
	logger.SetLevel(lvl)

	if strings.HasPrefix(idDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Fatalf("Cannot resolve home directory: %v", err)
		}
		idDir = filepath.Join(home, idDir[1:])
	}
	logger.Infof("Initializing DORD (Level: %s)", logLevel)

	s, err := server.New(addr, idDir, port)
	if err != nil {
		logger.Fatalf("Error initializing server: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := s.Serve(ctx); err != nil && err != context.Canceled {
		cmd.PrintErrln("Error running server:", err)
		os.Exit(1)
	}
	logger.Infof("Sutdown complete.")
}

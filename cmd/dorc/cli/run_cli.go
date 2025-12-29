package cli

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
)

func runCLI(c *client.Client) {
	done := make(chan struct{})

	go func() {
		for ev := range c.Events() {
			logEvent(ev)
		}
		close(done)
	}()

	c.Run()
	<-done
}

func logEvent(ev client.Event) {
	switch ev.Level {
	case logger.Debug:
		logger.LogDebug("%s", ev.Message)
	case logger.Info:
		logger.LogInfo("%s", ev.Message)
	case logger.Warn:
		logger.LogWarning("%s", ev.Message)
	case logger.Error:
		logger.LogError("%s", ev.Message)
	}
}

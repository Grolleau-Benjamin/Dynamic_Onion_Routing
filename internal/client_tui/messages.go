package clienttui

import "github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"

type eventMsg struct {
	Event client.Event
}

type doneMsg struct{}

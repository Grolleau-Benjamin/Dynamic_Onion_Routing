package clienttui

type Mode int

const (
	ModeHome      Mode = iota // Log messages
	ModeSelection             // 3 cases -> onion-path / dest / payload
)

type State struct {
	Width  int
	Height int

	Mode Mode
}

func NewState() State {
	return State{
		Mode: ModeHome,
	}
}

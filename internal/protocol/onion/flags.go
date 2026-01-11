package onion

const (
	FlagLastServer = 0x08 // 0000 1000
	FlagNbNextHops = 0x07 // 0000 0111
)

func IsLastServer(flags uint8) bool {
	return (flags & FlagLastServer) != 0
}

func GetNbNextHops(flags uint8) uint8 {
	return flags & FlagNbNextHops
}

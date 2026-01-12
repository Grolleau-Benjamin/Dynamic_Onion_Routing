package packet

type PacketFactory func() Packet

var registry = map[uint8]PacketFactory{
	TypeGetIdentityRequest:  func() Packet { return &GetIdentityRequest{} },
	TypeGetIdentityResponse: func() Packet { return &GetIdentityResponse{} },

	TypeOnionPacket: func() Packet { return &OnionPacket{} },
}

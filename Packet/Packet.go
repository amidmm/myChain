package Packet

import (
	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
)

func GetPacket(hash []byte) (*msg.Packet, error) {
	return &msg.Packet{}, Consts.ErrNotImplemented
}

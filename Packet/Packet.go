package Packet

import (
	"github.com/amidmm/MyChain/Account"
	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/golang/protobuf/proto"
)

func GetPacket(hash []byte) (*msg.Packet, error) {
	return &msg.Packet{}, Consts.ErrNotImplemented
}

func SetPacketSign(p *msg.Packet, u Account.User) error {
	sign, err := GetPacketSign(p, u)
	if err != nil {
		return err
	}
	p.Sign = sign
	return nil
}

func GetPacketSign(p *msg.Packet, u Account.User) ([]byte, error) {
	p.Sign = nil
	raw, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}
	sign, err := u.PrivKey.Sign(raw)
	if err != nil {
		return nil, err
	}
	return sign, nil
}

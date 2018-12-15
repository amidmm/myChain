package Packet

import (
	"github.com/amidmm/MyChain/Account"
	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/go-ethereum/crypto/sha3"
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

// SetHash sets the hash of passed packet
func SetHash(p *msg.Packet) error {
	hash, err := GetHash(*p)
	if err != nil {
		return err
	}
	p.Hash = hash
	return nil
}

func GetHash(p msg.Packet) ([]byte, error) {
	p.Hash = nil
	hash := sha3.New512()
	raw, err := proto.Marshal(&p)
	if err != nil {
		return nil, err
	}
	hash.Write(raw)
	return hash.Sum(nil), nil
}

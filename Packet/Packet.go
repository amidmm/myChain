package Packet

import (
	"github.com/golang/protobuf/jsonpb"

	"github.com/amidmm/MyChain/Account"
	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/go-ethereum/crypto/sha3"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-crypto"
)

func GetPacket(hash []byte) (*msg.Packet, error) {
	return &msg.Packet{}, Consts.ErrNotImplemented
}

func SetPacketSign(p *msg.Packet, u *Account.User) error {
	sign, err := GetPacketSign(p, u)
	if err != nil {
		return err
	}
	p.Sign = sign
	return nil
}

func GetPacketSign(p *msg.Packet, u *Account.User) ([]byte, error) {
	p.Sign = nil
	nonce := p.Nonce
	hash := p.Hash
	p.Nonce = nil
	p.Hash = nil
	defer func() {
		p.Nonce = nonce
		p.Hash = hash
	}()
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

func ValidateSign(p *msg.Packet) (bool, error) {
	userPK, err := crypto.UnmarshalPublicKey(p.Addr)
	if err != nil {
		return false, err
	}
	sign := p.Sign
	nonce := p.Nonce
	hash := p.Hash
	p.Sign = nil
	p.Nonce = nil
	p.Hash = nil
	defer func() {
		p.Sign = sign
		p.Nonce = nonce
		p.Hash = hash
	}()
	raw, err := proto.Marshal(p)
	if err != nil {
		return false, err
	}
	if ok, err := userPK.Verify(raw, sign); err != nil || !ok {
		return false, err
	}
	return true, nil
}

func ToJSON(p *msg.Packet) string {
	json := jsonpb.Marshaler{}
	s, _ := json.MarshalToString(p)
	return s
}

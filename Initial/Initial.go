package Initial

import (
	"bytes"
	"errors"

	"github.com/amidmm/MyChain/Account"

	"golang.org/x/crypto/sha3"

	"github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p-crypto"

	"github.com/amidmm/MyChain/Bundle"
	"github.com/amidmm/MyChain/Messages"
	"github.com/amidmm/MyChain/Tangle"
)

func ValidateInitial(p *msg.Packet, t *Tangle.Tangle) (bool, error) {
	if p.GetInitialData().PoBurn != nil {
		if ok, err := Bundle.ValidateBundle(p.GetInitialData().PoBurn, false); err != nil || !ok {
			return false, err
		}
		if ok := Bundle.HasBurn(p.GetInitialData().PoBurn); !ok {
			return false, errors.New("non-negative tx for Proof-of-burn")
		}
	}
	if p.GetInitialData().OwnerAddr != nil {
		u, _ := crypto.UnmarshalPublicKey(p.Addr)
		raw := bytes.Join([][]byte{
			p.CurrentBlockHash, p.Addr}, []byte{})
		if ok, err := u.Verify(raw, p.GetInitialData().OwnerAddrSign); err != nil || !ok {
			return false, err
		}
	}
	if p.GetInitialData().IpfsDetail != nil {
		_, err := multiaddr.NewMultiaddrBytes(p.GetInitialData().IpfsDetail)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func NewInitial(ServiceName string, nonce []byte, bun *msg.Bundle, UserOwnerAddr *Account.User, currentUser *Account.User, CurrentBlockHash []byte, ipfsAddr multiaddr.Multiaddr, verify1 []byte, verify2 []byte) (*msg.Initial, error) {
	inital := &msg.Initial{}
	var err error
	currentUserByte, _ := currentUser.PubKey.Bytes()
	if nonce == nil {
		inital.Service = []byte(ServiceName)
	} else {
		hash := sha3.New512()
		hash.Write(
			bytes.Join([][]byte{[]byte(ServiceName), nonce}, []byte{}))
	}
	if UserOwnerAddr != nil {
		inital.OwnerAddr, _ = UserOwnerAddr.PubKey.Bytes()
		inital.OwnerAddrSign, err = UserOwnerAddr.PrivKey.Sign(
			bytes.Join([][]byte{CurrentBlockHash, currentUserByte}, []byte{}))
		if err != nil {
			return nil, err
		}
	}
	if ipfsAddr != nil {
		inital.IpfsDetail = ipfsAddr.Bytes()
	}
	if bun != nil {
		if ok := Bundle.HasBurn(bun); !ok {
			return nil, errors.New("non-negative tx for Proof-of-burn")
		}
		inital.PoBurn = bun
	}
	inital.Verify1 = verify1
	inital.Verify2 = verify2
	return inital, nil
}

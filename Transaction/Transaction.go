package Transaction

import (
	"bytes"
	"fmt"

	"github.com/libp2p/go-libp2p-crypto"

	"github.com/amidmm/MyChain/Account"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/sha3"
)

func ValidateTx(t *msg.Tx, coinbase bool) (bool, error) {
	if bytes.Compare(t.Hash, GetTxHash(t)) != 0 {
		return false, nil
	}
	if coinbase {
		if t.RefTx != nil {
			fmt.Println("string")
			return false, nil
		} else if v, _ := ValidateTxSign(t); !v {
			return false, nil
		}
		return true, nil
	}
	return true, Consts.ErrNotImplemented
}

func GetTxHash(t *msg.Tx) []byte {
	exBundleHash := t.BundleHash
	exHash := t.Hash
	t.BundleHash = nil
	t.Hash = nil
	raw, _ := proto.Marshal(t)
	hash := sha3.New512()
	hash.Write(raw)
	t.BundleHash = exBundleHash
	t.Hash = exHash
	return hash.Sum(nil)
}

func SetTxSign(t *msg.Tx, u *Account.User) error {
	t.Sign = nil
	t.BundleHash = nil
	raw, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	sign, err := u.PrivKey.Sign(raw)
	if err != nil {
		return err
	}
	t.Sign = sign
	return nil
}

func ValidateTxSign(t *msg.Tx) (bool, error) {
	sign := t.Sign
	BundleHash := t.BundleHash
	t.Sign = nil
	t.BundleHash = nil
	defer func() {
		t.Sign = sign
		t.BundleHash = BundleHash
	}()
	refTx, err := GetUTXO(t.RefTx)
	pub, err := crypto.UnmarshalPublicKey(refTx.Sign)
	if err != nil {
		return false, err
	}
	raw, err := proto.Marshal(t)
	if err != nil {
		return false, err
	}
	if v, err := pub.Verify(raw, sign); err != nil || !v {
		return false, err
	} else {
		return true, nil
	}
}

func GetUTXO(hash []byte) (*msg.Tx, error) {
	return nil, Consts.ErrNotImplemented
}

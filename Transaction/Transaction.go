package Transaction

import (
	"bytes"
	"fmt"

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
	t.BundleHash = []byte{}
	t.Hash = []byte{}
	raw, _ := proto.Marshal(t)
	hash := sha3.New512()
	hash.Write(raw)
	t.BundleHash = exBundleHash
	t.Hash = exHash
	return hash.Sum(nil)
}

func ValidateTxSign(t *msg.Tx) (bool, error) {
	return true, Consts.ErrNotImplemented
}

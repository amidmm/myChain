package Bundle

import (
	"bytes"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/amidmm/MyChain/Transaction"
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/sha3"
)

func ValidateBundle(bun *msg.Bundle, inBlock bool) (bool, error) {
	// not implemented
	if inBlock {
		if bun.BundleType != msg.Bundle_POWERFUL {
			return false, nil
		}
		if bytes.Compare(bun.Hash, GetBundleHash(bun)) != 0 || bun.Transactions != nil {
			return false, nil
		}
	}
	if v, _ := ValidateVerify(bun); !v {
		return false, nil
	}
	return true, Consts.ErrNotImplemented
}

func GetBundleHash(bun *msg.Bundle) []byte {
	bun.Hash = nil
	hash := sha3.New512()
	tx := bun.Transactions
	bun.Transactions = nil
	for _, v := range tx {
		hash.Write(Transaction.GetTxHash(v))
	}
	raw, _ := proto.Marshal(bun)
	hash.Write(raw)
	bun.Transactions = tx
	return hash.Sum(nil)
}

func ValidateVerify(bun *msg.Bundle) (bool, error) {
	return true, Consts.ErrNotImplemented
}

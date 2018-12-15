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
	//inBlock for artifical bundels
	if inBlock {
		if bun.BundleType != msg.Bundle_POWERFUL {
			return false, nil
		}
		if bytes.Compare(bun.Hash, GetBundleHash(*bun)) != 0 || bun.Transactions != nil {
			return false, nil
		}
	}
	if !bytes.Equal(bun.Hash, GetBundleHash(*bun)) {
		return false, nil
	}
	sum := int64(0)
	for _, tx := range bun.GetTransactions() {
		if r, err := Transaction.ValidateTx(tx, false); err != nil || !r {
			return false, nil
		}
		sum += tx.Value
	}
	if sum != 0 {
		return false, nil
	}
	if bun.BundleType == msg.Bundle_WEAK {
		return true, nil
	}
	if v, _ := ValidateVerify(bun); !v {
		return false, nil
	}
	return true, nil
}

func GetBundleHash(bun msg.Bundle) []byte {
	bun.Hash = nil
	hash := sha3.New512()
	tx := bun.Transactions
	bun.Transactions = nil
	for _, v := range tx {
		hash.Write(Transaction.GetTxHash(v))
	}
	raw, _ := proto.Marshal(&bun)
	hash.Write(raw)
	bun.Transactions = tx
	return hash.Sum(nil)
}

func ValidateVerify(bun *msg.Bundle) (bool, error) {
	return true, Consts.ErrNotImplemented
}

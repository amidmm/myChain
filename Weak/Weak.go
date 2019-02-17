package Weak

import (
	"bytes"
	"encoding/binary"

	"github.com/amidmm/MyChain/Bundle"
	"github.com/amidmm/MyChain/Messages"
	"golang.org/x/crypto/sha3"
)

func GetWeakReqHash(r msg.WeakReq) []byte {
	r.Hash = nil
	hash := sha3.New512()
	var total []byte
	binary.LittleEndian.PutUint32(total, r.Total)
	hash.Write(total)
	hash.Write(Bundle.GetBundleHash(*r.GetBurn()))
	return hash.Sum(nil)
}

func ValidateWeakReq(r *msg.WeakReq) (bool, error) {
	if bytes.Compare(r.Hash, GetWeakReqHash(*r)) != 0 {
		return false, nil
	}
	if r.Burn == nil {
		return false, nil
	}
	if v, err := Bundle.ValidateBundle(r.Burn, false); err != nil || !v {
		return false, err
	}
	return true, nil
}

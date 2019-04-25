package Weak

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/amidmm/MyChain/Consts"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"

	"github.com/amidmm/MyChain/Bundle"
	"github.com/amidmm/MyChain/Messages"
	"golang.org/x/crypto/sha3"
)

var WeakReqDB *leveldb.DB = nil

func init() {
	if WeakReqDB == nil {
		s, err := storage.OpenFile(Consts.WeakReqPool, false)
		if err != nil {
			log.Fatalln("unable to open WeakReqDB")
		}
		db, err := leveldb.Open(s, nil)
		if err != nil {
			log.Fatalln("unable to open WeakReqDB")
		}
		WeakReqDB = db
	}
}

func AddNewWeakReq(weak *msg.Packet, block *msg.Packet) bool {
	err := WeakReqDB.Put(weak.Addr, block.Addr, nil)
	if err != nil {
		return false
	}
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, weak.GetWeakData().TotalTx)
	err = WeakReqDB.Put(
		bytes.Join([][]byte{
			[]byte("num"), weak.Addr}, []byte{}), b, nil)
	if err != nil {
		return false
	}
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, weak.GetWeakData().TotalFee/weak.GetWeakData().TotalTx)
	err = WeakReqDB.Put(
		bytes.Join([][]byte{
			[]byte("fee"), weak.Addr}, []byte{}), b, nil)
	if err != nil {
		return false
	}
	return true
}

func CheckHasWeak(p *msg.Packet) bool {
	_, err := WeakReqDB.Get(p.Addr, nil)
	if err != nil {
		return false
	}
	return true
}

func WeakReqMinusMinus(p *msg.Packet) bool {
	addr, err := WeakReqDB.Get(p.Addr, nil)
	if err != nil {
		return false
	}
	bNum, err := WeakReqDB.Get(
		bytes.Join([][]byte{
			[]byte("num"), p.Addr}, []byte{}), nil)
	if err != nil {
		return false
	}
	num := binary.LittleEndian.Uint32(bNum)
	num--
	if num <= 0 {
		err = WeakReqDB.Delete(p.Addr, nil)
		if err != nil {
			return false
		}
	}

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, num)
	err = WeakReqDB.Put(
		bytes.Join([][]byte{
			[]byte("num"), p.Addr}, []byte{}), b, nil)
	if err != nil {
		return false
	}

	bFee, err := WeakReqDB.Get(
		bytes.Join([][]byte{
			[]byte("fee"), p.Addr}, []byte{}), nil)
	if err != nil {
		return false
	}
	fee := binary.LittleEndian.Uint32(bFee)
	if p.GetBundleData() != nil {
		exists := false
		for _, tx := range p.GetBundleData().Transactions {
			if bytes.Equal(tx.Sign, addr) && tx.Value > int64(fee) {
				exists = true
			}
		}
		if !exists {
			return false
		}
	} else {
		return false
	}
	return true
}

func GetWeakReqHash(r msg.WeakReq) []byte {
	r.Hash = nil
	hash := sha3.New512()
	var TotalFee = make([]byte, 4)
	binary.LittleEndian.PutUint32(TotalFee, r.TotalFee)
	hash.Write(TotalFee)
	var TotalTx = make([]byte, 4)
	binary.LittleEndian.PutUint32(TotalTx, r.TotalTx)
	hash.Write(TotalTx)
	hash.Write(Bundle.GetBundleHash(*r.GetBurn()))
	return hash.Sum(nil)
}

func ValidateWeakReq(x *msg.Packet) (bool, error) {
	r := x.GetWeakData()
	if bytes.Compare(r.Hash, GetWeakReqHash(*r)) != 0 {
		return false, nil
	}
	if r.Burn == nil {
		return false, nil
	}
	if v, err := Bundle.ValidateBundle(r.Burn, false); err != nil || !v {
		return false, err
	}
	if !Bundle.HasBurn(r.Burn) {
		return false, Consts.ErrNotABurn
	}
	return true, nil
}

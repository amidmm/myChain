package Transaction

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"sync"

	"github.com/syndtr/goleveldb/leveldb/storage"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/libp2p/go-libp2p-crypto"

	"github.com/amidmm/MyChain/Account"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/sha3"
)

var UTXOlock sync.Mutex
var UTXOdatabase *leveldb.DB
var ThisUserDb *leveldb.DB
var ThisUserAddr []byte
var LockMoney *leveldb.DB

func init() {
	err := OpenUTXO()
	if err != nil {
		log.Println("\033[31m TX: unable to open UTXO database " + err.Error() + "\033[0m")
	}
	if LockMoney == nil {
		s, err := storage.OpenFile(Consts.LockMoney, false)
		if err != nil {
			log.Fatalln("unable to open LockMoneyDB")
		}
		db, err := leveldb.Open(s, nil)
		if err != nil {
			log.Fatalln("unable to open LockMoneyDB")
		}
		LockMoney = db
	}
}

func OpenUTXO() error {
	//needed to change if multiple blockchain allowed
	UTXOlock.Lock()
	defer UTXOlock.Unlock()
	s, err := storage.OpenFile(Consts.UTXODB, false)
	if err != nil {
		return err
	}
	db, err := leveldb.Open(s, nil)
	if err != nil {
		return err
	}
	UTXOdatabase = db
	return nil
}

func ValidateTx(t *msg.Tx, coinbase bool) (bool, error) {
	if bytes.Compare(t.Hash, GetTxHash(*t)) != 0 {
		return false, nil
	}
	if coinbase {
		if t.RefTx != nil && t.BundleHash != nil {
			return false, nil
		}
		return true, nil
	}
	if t.Value < 0 {
		if v, err := ValidateTxSign(t); err != nil || !v {
			return false, err
		}
	} else {
		if t.Sign == nil {
			return false, nil
		}
	}
	return true, nil
}

func GetTxHash(t msg.Tx) []byte {
	t.BundleHash = nil
	t.Hash = nil
	t.Sign = nil
	raw, _ := proto.Marshal(&t)
	hash := sha3.New512()
	hash.Write(raw)
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
	//TODO: make it base on address
	//		"public key is not embedded in peer ID" error while using addr
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
	raw, err := UTXOdatabase.Get(hash, nil)
	if err != nil {
		return nil, err
	}
	t := &msg.Tx{}
	err = proto.Unmarshal(raw, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func PutUTXO(t *msg.Tx, addr []byte) error {
	raw, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	hash := GetTxHash(*t)
	err = UTXOdatabase.Put(hash, raw, nil)
	if err != nil {
		return err
	}
	if addr != nil && ThisUserDb != nil && bytes.Equal(addr, ThisUserAddr) {
		err = ThisUserDb.Put(hash, raw, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func UnUTXO(t msg.Tx) error {
	UTXOlock.Lock()
	defer UTXOlock.Unlock()
	hash := GetTxHash(t)
	err := UTXOdatabase.Delete(hash, nil)
	if err != nil {
		return err
	}
	if ThisUserDb != nil {
		ThisUserDb.Delete(hash, nil)
	}
	return nil
}

func UnUTXOWithHash(hash []byte) error {
	UTXOlock.Lock()
	defer UTXOlock.Unlock()
	err := UTXOdatabase.Delete(hash, nil)
	if err != nil {
		return err
	}
	if ThisUserDb != nil {
		ThisUserDb.Delete(hash, nil)
	}
	return nil
}

func CloseUTXO() error {
	UTXOlock.Lock()
	defer UTXOlock.Unlock()
	err := UTXOdatabase.Close()
	if err != nil {
		return err
	}
	log.Println("UTXO database is down...")
	return nil
}

func HandleBundle(p *msg.Packet) (bool, error) {
	last := 0
	var Transactions []*msg.Tx
	if p.PacketType == msg.Packet_BUNDLE {
		Transactions = p.GetBundleData().Transactions
	} else if p.PacketType == msg.Packet_WEAKREQ {
		Transactions = p.GetWeakData().Burn.Transactions
	} else if p.PacketType == msg.Packet_INITIAL {
		if p.GetInitialData().PoBurn == nil {
			return true, nil
		}
		Transactions = p.GetInitialData().PoBurn.Transactions
	} else {
		return true, nil
	}
	defer func() {
		if Transactions[last] != Transactions[len(Transactions)-1] {
			for ; last >= 0; last-- {
				if Transactions[last].Value > 0 {
					UnUTXOWithHash(Transactions[last].Hash)
				} else if Transactions[last].Value < 0 {
					PutUTXO(Transactions[last], p.Addr)
				}
			}
		}
	}()
	data := Transactions
	for i, v := range data {
		if v.Value > 0 {
			if err := PutUTXO(v, p.Addr); err != nil {
				return false, err
			}
		} else if v.Value < 0 {
			if err := UnUTXOWithHash(v.Hash); err != nil {
				return false, err
			}
		}
		last = i
	}
	return true, nil
}

func HandleLockBundle(p *msg.Packet) (bool, error) {
	last := 0
	var Transactions []*msg.Tx
	var Addrs [][]byte
	Transactions = p.GetRepData().GetAgreeData().LockMoney.Transactions
	Addrs = append(Addrs, p.GetRepData().GetAgreeData().Mediator)
	Addrs = append(Addrs, p.Addr)
	defer func() {
		if Transactions[last] != Transactions[len(Transactions)-1] {
			for ; last >= 0; last-- {
				if Transactions[last].Value > 0 {
					UnUTXOWithHash(Transactions[last].Hash)
				} else if Transactions[last].Value < 0 {
					PutLockMoney(Transactions[last], Addrs)
				}
			}
		}
	}()
	data := Transactions
	for i, v := range data {
		if v.Value > 0 {
			if err := PutLockMoney(v, Addrs); err != nil {
				return false, err
			}
		} else if v.Value < 0 {
			if err := UnUTXOWithHash(v.Hash); err != nil {
				return false, err
			}
		}
		last = i
	}
	return true, nil
}

func PutLockMoney(t *msg.Tx, UnLocker [][]byte) error {
	raw, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	hash := GetTxHash(*t)
	err = LockMoney.Put(hash, raw, nil)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(UnLocker)
	err = LockMoney.Put(bytes.Join(
		[][]byte{[]byte("Unlocker"), hash}, []byte{}), buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

func UnLockMoney(hash []byte, addr []byte) error {
	raw, err := LockMoney.Get(bytes.Join(
		[][]byte{[]byte("Unlocker"), hash}, []byte{}), nil)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.Write(raw)
	var addrs [][]byte
	dec := gob.NewDecoder(&buf)
	dec.Decode(addrs)
	IsValid := false
	for _, a := range addrs {
		if bytes.Equal(a, addr) {
			IsValid = true
		}
	}
	if !IsValid {
		return errors.New("not a right person to unlock")
	}
	raw, err = LockMoney.Get(hash, nil)
	if err != nil {
		return err
	}
	LockMoney.Delete(hash, nil)
	LockMoney.Delete(bytes.Join(
		[][]byte{[]byte("Unlocker"), hash}, []byte{}), nil)

	tx := &msg.Tx{}
	proto.Unmarshal(raw, tx)
	PutUTXO(tx, tx.Sign)
	return nil
}

package Transaction

import (
	"bytes"
	"fmt"
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

func PutUTXO(t *msg.Tx) error {
	raw, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	hash := GetTxHash(t)
	err = UTXOdatabase.Put(hash, raw, nil)
	if err != nil {
		return err
	}
	return nil
}

func UnUTXO(t *msg.Tx) error {
	UTXOlock.Lock()
	defer UTXOlock.Unlock()
	hash := GetTxHash(t)
	err := UTXOdatabase.Delete(hash, nil)
	if err != nil {
		return err
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

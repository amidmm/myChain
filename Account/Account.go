package Account

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/amidmm/MyChain/Messages"

	"github.com/libp2p/go-libp2p-peer"

	"github.com/amidmm/MyChain/Consts"
	"github.com/syndtr/goleveldb/leveldb/storage"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/libp2p/go-libp2p-crypto"
)

type User struct {
	Name     string
	PubKey   crypto.PubKey
	PrivKey  crypto.PrivKey
	Addr     peer.ID
	userLock sync.Mutex
	DB       *leveldb.DB
}

func CreateUser(name string, pub crypto.PubKey, priv crypto.PrivKey, debug bool, debugSeed int64) error {
	u := &User{}
	u.userLock.Lock()
	defer u.userLock.Unlock()
	u.Name = name
	if debug {
		if err := u.setKey(debugSeed); err != nil {
			return err
		}
	} else if pub != nil && priv != nil {
		u.PubKey = pub
		u.PrivKey = priv
	} else {
		u.setKey(0)
	}
	err := SetAddr(u)
	if err != nil {
		return err
	}
	path, err := storage.OpenFile(fmt.Sprintf(Consts.UserDB, name), false)
	if err != nil {
		return err
	}
	defer path.Close()
	db, err := leveldb.Open(path, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	raw, err := crypto.MarshalPublicKey(u.PubKey)
	if err != nil {
		return err
	}
	db.Put([]byte("metadatapub"), raw, nil)
	raw, err = crypto.MarshalPrivateKey(u.PrivKey)
	if err != nil {
		return err
	}
	db.Put([]byte("metadatapriv"), raw, nil)
	db.Put([]byte("metadataaddr"), []byte(peer.IDB58Encode(u.Addr)), nil)
	return nil
}

func (u *User) setKey(seed int64) error {
	var r io.Reader
	if seed != 0 {
		r = mrand.New(mrand.NewSource(int64(seed)))
	} else {
		r = rand.Reader
	}
	priv, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 1024, r)
	if err != nil {
		return err
	}
	u.PrivKey = priv
	u.PubKey = pubKey
	return nil
}

func LoadUser(name string) (*User, error) {
	u := &User{}
	u.Name = name
	path, err := storage.OpenFile(fmt.Sprintf(Consts.UserDB, name), false)
	if err != nil {
		return nil, err
	}
	db, err := leveldb.Open(path, nil)
	if err != nil {
		path.Close()
		return nil, err
	}
	rawPub, err := db.Get([]byte("metadatapub"), nil)
	if err != nil {
		path.Close()
		db.Close()
		return nil, err
	}
	if u.PubKey, err = crypto.UnmarshalPublicKey(rawPub); err != nil {
		return nil, err
	}
	rawPriv, err := db.Get([]byte("metadatapriv"), nil)
	if err != nil {
		return nil, err
	}
	if u.PrivKey, err = crypto.UnmarshalPrivateKey(rawPriv); err != nil {
		return nil, err
	}
	rawAddr, err := db.Get([]byte("metadataaddr"), nil)
	if err != nil {
		return nil, err
	}
	rawStringAddr := string(rawAddr)
	u.Addr, err = peer.IDB58Decode(rawStringAddr)
	if err != nil {
		return nil, err
	}
	u.DB = db
	return u, nil
}

func SetAddr(u *User) error {
	var err error
	u.Addr, err = peer.IDFromPublicKey(u.PubKey)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) GetUTXOList() []*msg.Tx {
	iter := u.DB.NewIterator(nil, nil)
	var list []*msg.Tx
	for iter.Next() {
		if bytes.HasPrefix(iter.Key(), []byte("metadata")) {
			continue
		}
		packet := &msg.Tx{}
		_ = proto.Unmarshal(iter.Value(), packet)
		list = append(list, packet)
	}
	iter.Release()
	return list
}

func (u *User) GetUTXO(value int64) *msg.Tx {
	tx := &msg.Tx{}
	if value <= 0 {
		return nil
	}
	iter := u.DB.NewIterator(nil, nil)
	for iter.Next() {
		if bytes.HasPrefix(iter.Key(), []byte("metadata")) {
			continue
		}
		packet := &msg.Tx{}
		_ = proto.Unmarshal(iter.Value(), packet)
		if packet.Value >= value {
			tx = packet
			break
		}
	}
	iter.Release()
	return tx
}

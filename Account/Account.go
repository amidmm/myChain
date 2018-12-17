package Account

import (
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"sync"

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
	db.Put([]byte("pub"), raw, nil)
	raw, err = crypto.MarshalPrivateKey(u.PrivKey)
	if err != nil {
		return err
	}
	db.Put([]byte("priv"), raw, nil)
	db.Put([]byte("addr"), []byte(peer.IDB58Encode(u.Addr)), nil)
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
	defer path.Close()
	db, err := leveldb.Open(path, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rawPub, err := db.Get([]byte("pub"), nil)
	if err != nil {
		return nil, err
	}
	if u.PubKey, err = crypto.UnmarshalPublicKey(rawPub); err != nil {
		return nil, err
	}
	rawPriv, err := db.Get([]byte("priv"), nil)
	if err != nil {
		return nil, err
	}
	if u.PrivKey, err = crypto.UnmarshalPrivateKey(rawPriv); err != nil {
		return nil, err
	}
	rawAddr, err := db.Get([]byte("addr"), nil)
	if err != nil {
		return nil, err
	}
	rawStringAddr := string(rawAddr)
	u.Addr, err = peer.IDB58Decode(rawStringAddr)
	if err != nil {
		return nil, err
	}
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

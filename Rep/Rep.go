package Rep

import (
	"log"

	"github.com/go-ethereum/crypto/sha3"
	"github.com/golang/protobuf/proto"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

//change statistics
var RepPool *leveldb.DB = nil

func init() {
	if RepPool == nil {
		s, err := storage.OpenFile(Consts.RepPool, false)
		if err != nil {
			log.Fatalln("unable to open WeakReqDB")
		}
		db, err := leveldb.Open(s, nil)
		if err != nil {
			log.Fatalln("unable to open WeakReqDB")
		}
		RepPool = db
	}
}

func GetHash(rep *msg.Rep) []byte {
	rep.Hash = nil
	raw, err := proto.Marshal(rep)
	if err == nil {
		return nil
	}
	hash := sha3.New512()
	hash.Write(raw)
	return hash.Sum(nil)
}

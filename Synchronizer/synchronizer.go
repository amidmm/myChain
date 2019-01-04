package Synchronizer

import (
	"github.com/amidmm/MyChain/Blockchain"
	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/amidmm/MyChain/Tangle"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

var UnsyncPoolDB *leveldb.DB

func init() {
	s, err := storage.OpenFile(Consts.UnsyncPool, false)
	if err != nil {
		return
	}
	UnsyncPoolDB, err = leveldb.Open(s, nil)
	if err != nil {
		return
	}
}

func HasBeenSeen(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) bool {
	switch p.Data.(type) {
	case *msg.Packet_BlockData:
		return bc.HasBlockSeen(p)
	default:
		return t.HasSeenInTangle(p)
	}
}

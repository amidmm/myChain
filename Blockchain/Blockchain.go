package Blockchain

import (
	"bytes"
	"context"
	"errors"

	"github.com/syndtr/goleveldb/leveldb/storage"

	"github.com/golang/protobuf/ptypes"

	"github.com/amidmm/MyChain/PoW"

	"github.com/golang/protobuf/proto"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/syndtr/goleveldb/leveldb"
)

type Blockchain struct {
	Tip *msg.Packet
	DB  *leveldb.DB
}
type BlockchainIterator struct {
	Tip msg.Packet
	Err error
}

func OpenBlockChain() (*leveldb.DB, error) {
	s, err := storage.OpenFile(Consts.BlockchainDB, false)
	if err != nil {
		return nil, err
	}
	db, err := leveldb.Open(s, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

//NewBlockchain creates a new blockchain
func NewBlockchain() (*Blockchain, error) {
	db, err := OpenBlockChain()
	if err != nil {
		return nil, err
	}
	_, err = db.Get([]byte("l"), nil)
	if err == leveldb.ErrNotFound {
		b := Blockchain{}
		genesis := GenesisBlock()
		//TODO: use add block instead
		b.Tip = genesis
		raw, _ := proto.Marshal(genesis)
		err = db.Put(
			bytes.Join([][]byte{
				[]byte("b"), genesis.Hash}, []byte{}), raw, nil)
		if err != nil {
			return nil, err
		}
		err = db.Put([]byte("l"), genesis.Hash, nil)
		if err != nil {
			return nil, err
		}
		b.DB = db
		return &b, nil
	}
	return nil, Consts.ErrBlockchainExists
}

//ReadBlockchainTip reads the last block of blockchain
func (b *Blockchain) ReadBlockchainTip() (*msg.Packet, error) {
	db := b.DB
	value, err := db.Get([]byte("l"), nil)
	if err != nil {
		return nil, err
	}
	rawBlock, err := db.Get(bytes.Join(
		[][]byte{[]byte("b"), value}, []byte{}),
		nil)
	if err != nil {
		return nil, err
	}
	block := &msg.Packet{}
	err = proto.Unmarshal(rawBlock, block)
	if err != nil {
		return nil, err
	}
	switch block.Data.(type) {
	case *msg.Packet_BlockData:
		return block, nil
	default:
		panic(errors.New("wrong Packet type as last block"))
	}
}

//GenesisBlock generates the first block
func GenesisBlock() *msg.Packet {
	//TODO: change addrs, Sign and Bundle to a valid one
	packet := &msg.Packet{}
	packet.Addr = []byte{}
	packet.CurrentBlockNumber = 1
	packet.Diff = 1
	packet.PacketType = msg.Packet_BLOCK
	packet.Prev = []byte{}
	packet.Sign = []byte("This is the genesis")
	packet.Timestamp = ptypes.TimestampNow()
	block := &msg.Block{}
	block.Reqs = []*msg.WeakReq{}
	block.Sanities = []*msg.SanityCheck{}
	block.BundleHashs = []*msg.HashArray{}
	block.Coinbase = &msg.Tx{}
	packet.Data = &msg.Packet_BlockData{block}
	PoW.SetPoW(context.Background(), packet, 1)
	PoW.SetHash(packet)
	return packet
}

// AddBlock add new block to current blockchain
// Tip: validation should get done before hand
func (b *Blockchain) AddBlock(p *msg.Packet) error {
	db := b.DB
	raw, _ := proto.Marshal(p)
	err := db.Put(
		bytes.Join([][]byte{
			[]byte("b"), p.Hash}, []byte{}), raw, nil)
	if err != nil {
		return err
	}
	err = db.Put([]byte("l"), p.Hash, nil)
	if err != nil {
		return err
	}
	return nil
}

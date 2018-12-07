package Blockchain

import (
	"bytes"
	"context"
	"errors"

	"github.com/golang/protobuf/ptypes"

	"github.com/amidmm/MyChain/PoW"

	"github.com/golang/protobuf/proto"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/syndtr/goleveldb/leveldb"
)

type blockchain struct {
	Tip *msg.Packet
}

func NewBlockchain() (*blockchain, error) {
	db, err := leveldb.OpenFile(Consts.BlockchainDB, nil)
	if err != nil {
		return nil, err
	}
	_, err = db.Get([]byte("l"), nil)
	if err == leveldb.ErrNotFound {
		b := blockchain{}
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
		return &b, nil
	}
	return nil, Consts.ErrBlockchainExists
}

//Reads the last block of blockchain
func ReadBlockchainTip() (*msg.Packet, error) {
	db, err := leveldb.OpenFile(Consts.BlockchainDB, nil)
	defer db.Close()
	if err != nil {
		return nil, err
	}
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
		return nil, errors.New("wrong Packet type as last block")
	}
}

//Generate the first block
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

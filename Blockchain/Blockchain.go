package Blockchain

import (
	"bytes"
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/amidmm/MyChain/Utils"

	"github.com/golang/protobuf/jsonpb"

	"github.com/syndtr/goleveldb/leveldb/storage"

	"github.com/amidmm/MyChain/Packet"
	"github.com/amidmm/MyChain/PoW"
	"github.com/amidmm/MyChain/Transaction"

	"github.com/golang/protobuf/proto"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/syndtr/goleveldb/leveldb"
)

type Blockchain struct {
	Tip                 *msg.Packet
	DB                  *leveldb.DB
	Relation            *leveldb.DB
	chainLock           sync.Mutex
	minRetargetTimespan int64  // target timespan / adjustment factor
	maxRetargetTimespan int64  // target timespan * adjustment factor
	blocksPerRetarget   uint64 // target timespan / target time per block
	chainParams         *chainParams
	GenesisHash         []byte
}

type chainParams struct {
	CoinbaseMaturity         uint16
	TargetTimespan           time.Duration
	TargetTimePerBlock       time.Duration
	RetargetAdjustmentFactor int64
	ReduceMinDifficulty      bool
	MinDiffReductionTime     time.Duration
	GenerateSupported        bool
}

func (b *Blockchain) InitBlockchain() {
	params := chainParams{
		CoinbaseMaturity:         100,
		TargetTimespan:           time.Hour * 24 * 14, // 14 days
		TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
		RetargetAdjustmentFactor: 2,                   // 25% less, 400% more
		ReduceMinDifficulty:      false,
		MinDiffReductionTime:     0,
		GenerateSupported:        false,
	}
	targetTimespan := int64(params.TargetTimespan / time.Second)         //TargetTimespan in seconds
	targetTimePerBlock := int64(params.TargetTimePerBlock / time.Second) //TargetTimePerBlock in seconds
	adjustmentFactor := params.RetargetAdjustmentFactor
	b.blocksPerRetarget = uint64(targetTimespan / targetTimePerBlock) //2016
	b.minRetargetTimespan = targetTimespan / adjustmentFactor         // 1209600÷4= 302400
	b.maxRetargetTimespan = targetTimespan * adjustmentFactor         // 1209600x4= 4838400
	b.chainParams = &params
}

type BlockchainIterator struct {
	Tip      msg.Packet
	DB       *leveldb.DB
	Relation *leveldb.DB
	Err      error
}

var DataBase *leveldb.DB = nil
var Relation *leveldb.DB = nil
var genericLock sync.Mutex

// OpenBlockChain opens blockchain Database
func OpenBlockChain() (*leveldb.DB, *leveldb.DB, error) {
	if DataBase != nil && Relation != nil {
		return DataBase, Relation, nil
	}
	genericLock.Lock()
	defer genericLock.Unlock()
	s, err := storage.OpenFile(Consts.BlockchainDB, false)
	if err != nil {
		return nil, nil, err
	}
	db, err := leveldb.Open(s, nil)
	if err != nil {
		return nil, nil, err
	}
	re, err := storage.OpenFile(Consts.BlockchainRelation, false)
	if err != nil {
		return nil, nil, err
	}
	rel, err := leveldb.Open(re, nil)
	if err != nil {
		return nil, nil, err
	}
	DataBase = db
	Relation = rel
	return db, rel, nil
}

//NewBlockchain creates a new blockchain
func NewBlockchain() (*Blockchain, error) {
	db, rel, err := OpenBlockChain()
	if err != nil {
		return nil, err
	}
	_, err = db.Get([]byte("l"), nil)
	if err == leveldb.ErrNotFound {
		b := Blockchain{}
		genesis := GenesisBlock()
		//TODO: use add block instead
		b.Tip = genesis
		b.GenesisHash = genesis.Hash
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
		err = rel.Put(genesis.Hash, []byte{}, nil)
		if err != nil {
			return nil, err
		}
		b.DB = db
		b.Relation = rel
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
		panic(Consts.ErrNotABlock)
	}
}

//GenesisBlock generates the first block
func GenesisBlock() *msg.Packet {
	packet := &msg.Packet{}
	packet.Addr = []byte{}
	packet.CurrentBlockNumber = 1
	packet.PacketType = msg.Packet_BLOCK
	packet.Prev = []byte{}
	packet.Sign = []byte("This is the genesis")
	block := &msg.Block{}
	block.Reqs = []*msg.Packet{}
	block.Sanities = []*msg.SanityCheck{}
	block.PacketHashs = []*msg.HashArray{}
	block.Coinbase = &msg.Tx{}
	packet.Data = &msg.Packet_BlockData{block}
	PoW.SetPoW(context.Background(), packet, 1)
	packet.Diff = Consts.PoWLimit
	Packet.SetHash(packet)
	return packet
}

// AddBlock add new block to current blockchain
// Tip: validation should get done before hand
func (b *Blockchain) AddBlock(p *msg.Packet) error {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	// TODO: change the expected target
	db := b.DB
	_, err := db.Get(p.Hash, nil)
	if err == leveldb.ErrNotFound && p.CurrentBlockNumber > b.Tip.CurrentBlockNumber {
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
		Transaction.OpenUTXO()
		if err != nil {
			return err
		}
		Transaction.PutUTXO(p.GetBlockData().Coinbase, p.Addr)
		if err != nil {
			return err
		}

		rawPrev, err := b.Relation.Get(p.Prev, nil)
		if err == leveldb.ErrNotFound {
			err = nil
			rawPrev = []byte{}
		}
		if err != nil {
			return err
		}
		rawPrev, err = Utils.AppendMarshalSha3(rawPrev, p.Hash)
		if err != nil {
			return err
		}
		b.Relation.Put(p.Prev, rawPrev, nil)

		b.Tip = p
		return nil
	}
	return Consts.ErrOldBlock
}

func (bc *Blockchain) CheckWithTip(p *msg.Packet) bool {
	//TODO: make it base on the packetHashes, rep, sanity
	if p.CurrentBlockNumber > bc.Tip.CurrentBlockNumber {
		return true
	}
	return false
}

//Prev returns the prev block of the BlockchainIterator
func (b *BlockchainIterator) Prev() bool {
	if b.Tip.Prev == nil {
		return false
	}
	db := b.DB
	rawBlock, err := db.Get(bytes.Join(
		[][]byte{[]byte("b"), b.Tip.Prev}, []byte{}),
		nil)
	if err != nil {
		b.Err = err
		return false
	}
	r := &msg.Packet{}
	err = proto.Unmarshal(rawBlock, r)
	if err != nil {
		b.Err = err
		return false
	}
	b.Tip = *r
	return true
}

// Value retrive the current value for iterator
func (b *BlockchainIterator) Value() (*msg.Packet, error) {
	if b.Err != nil {
		return nil, b.Err
	}
	return &b.Tip, nil
}

// ResetErr rests error value for iterator
func (b *BlockchainIterator) ResetErr() {
	b.Err = nil
}

// InitIter initalize the iterator for a blockchain
func (b *BlockchainIterator) InitIter(blockchain *Blockchain) error {
	if blockchain.DB == nil && blockchain.Relation == nil {
		db, rel, err := OpenBlockChain()
		if err != nil {
			b.Err = err
			return err
		}
		blockchain.DB = db
		blockchain.Relation = rel
		b.DB = db
		b.Relation = rel
	}
	b.DB = blockchain.DB
	b.Relation = blockchain.Relation
	b.Tip = *blockchain.Tip
	return nil
}

func (b *BlockchainIterator) Seek(blockHash []byte) error {
	tipraw, err := b.DB.Get(bytes.Join(
		[][]byte{[]byte("b"), blockHash}, []byte{}),
		nil)
	if err != nil {
		return err
	}
	tip := &msg.Packet{}
	err = proto.Unmarshal(tipraw, tip)
	b.Tip = *tip
	return err
}

func (b *BlockchainIterator) Next() bool {
	raw, err := b.Relation.Get(b.Tip.Hash, nil)
	if err != nil {
		b.Err = err
		return false
	}
	sha3Out, err := Utils.UnMarshalSha3List(raw)
	if err != nil {
		return false
	}
	if len(sha3Out) == 0 {
		return false
	}
	tip := &msg.Packet{}
	data, _ := b.DB.Get(bytes.Join([][]byte{[]byte("b"), sha3Out[0]}, []byte{}), nil)
	_ = proto.Unmarshal(data, tip)
	b.Tip = *tip
	return true
}

//ExportToJSON exports Blockchain as JSON
func (b *Blockchain) ExportToJSON(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bytes.NewBuffer([]byte{})
	_, err = buf.WriteString("[")
	if err != nil {
		return err
	}
	m := jsonpb.Marshaler{}
	iter := &BlockchainIterator{}
	err = iter.InitIter(b)
	if err != nil {
		return err
	}
	v, err := iter.Value()
	if err != nil {
		return err
	}
	err = m.Marshal(buf, v)
	if err != nil {
		return err
	}
	for iter.Prev() {
		_, err = buf.WriteString(",")
		if err != nil {
			return err
		}
		v, err := iter.Value()
		if err != nil {
			return err
		}
		err = m.Marshal(buf, v)
		if err != nil {
			return err
		}
	}
	_, err = buf.WriteString("]")
	if err != nil {
		return err
	}
	_, err = f.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (b *Blockchain) ReadBlock(hash []byte) (*msg.Packet, error) {
	db := b.DB
	rawBlock, err := db.Get(bytes.Join(
		[][]byte{[]byte("b"), hash}, []byte{}),
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
		panic(Consts.ErrNotABlock)
	}
}

func (b *Blockchain) Ancestor(height uint64) (*msg.Packet, error) {
	if height < 1 || height > b.Tip.CurrentBlockNumber {
		return nil, Consts.ErrWrongHeight
	}
	n := b.Tip
	for ; n != nil && n.CurrentBlockNumber != height; n, _ = b.ReadBlock(n.Prev) {
		// Intentionally left blank
	}
	return n, nil
}

func (b *Blockchain) RelativeAncestor(p *msg.Packet, distance uint64) (*msg.Packet, error) {
	return b.Ancestor(p.CurrentBlockNumber - distance)
}

func (b *Blockchain) CalcNextRequiredDifficulty() (uint32, error) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	if (b.Tip.CurrentBlockNumber+1)%b.NextRetarget() != 0 {
		return b.Tip.Diff, nil
	}
	firstNode, err := b.RelativeAncestor(b.Tip, b.blocksPerRetarget-2)
	if err != nil {
		return 0, Consts.ErrRetargetRetriv
	}
	actualTimespan := b.Tip.Timestamp.Seconds - firstNode.Timestamp.Seconds
	adjustedTimespan := actualTimespan
	if actualTimespan < b.minRetargetTimespan {
		adjustedTimespan = b.minRetargetTimespan
	} else if actualTimespan > b.maxRetargetTimespan {
		adjustedTimespan = b.maxRetargetTimespan
	}
	targetTimeSpan := int64(b.chainParams.TargetTimespan / time.Second)
	diff := int64(b.Tip.Diff) * (targetTimeSpan / adjustedTimespan)
	if uint32(diff) < Consts.PoWLimit {
		return Consts.PoWLimit, nil
	}
	return uint32(diff), nil
}

func (b *Blockchain) NextRetarget() uint64 {
	retarget := b.blocksPerRetarget - (b.Tip.CurrentBlockNumber % b.blocksPerRetarget)
	return retarget + b.Tip.CurrentBlockNumber
}

func (b *Blockchain) ExpectedBlockReward() (int64, error) {
	// 9 223 372,036 854 775 807
	// 1 000 000,000 000 000 001
	//TODO: should be dynamic
	return 100000000000, Consts.ErrNotImplemented
}

func (b *Blockchain) Close() error {
	if b.DB == nil || DataBase == nil {
		return leveldb.ErrClosed
	}
	b.chainLock.Lock()
	DataBase = nil
	b.DB.Close()
	log.Println("Blockchain database is down...")
	b.chainLock.Unlock()
	return nil
}

func (bc *Blockchain) HasBlockSeen(block *msg.Packet) bool {
	_, err := bc.ReadBlock(block.Hash)
	if err == nil {
		return true
	}
	return false
}

package Tangle

import (
	"bytes"
	"os"
	"sync"
	"time"

	"github.com/amidmm/MyChain/Utils"

	"github.com/amidmm/MyChain/Packet"

	"github.com/amidmm/MyChain/Blockchain"
	"github.com/amidmm/MyChain/Bundle"
	"github.com/golang/protobuf/ptypes"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type Tangle struct {
	Relations           *leveldb.DB
	UnApproved          *leveldb.DB
	DB                  *leveldb.DB
	Blockchain          *Blockchain.Blockchain
	tangleLock          sync.Mutex
	minRetargetTimespan int64  // target timespan / adjustment factor
	maxRetargetTimespan int64  // target timespan * adjustment factor
	blocksPerRetarget   uint64 // target timespan / target time per block
	TangleParams        *tanleParams
}

type tanleParams struct {
	TargetTimespan           time.Duration
	TargetTimePerBlock       time.Duration
	LastBlockToVerify        uint64
	RetargetAdjustmentFactor int64
	ReduceMinDifficulty      bool
	MinDiffReductionTime     time.Duration
	GenerateSupported        bool
	MaxBlockDiff             uint64
}

func (t *Tangle) InitTangle() {
	params := tanleParams{
		LastBlockToVerify:        100,
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
	t.blocksPerRetarget = uint64(targetTimespan / targetTimePerBlock) //2016
	t.minRetargetTimespan = targetTimespan / adjustmentFactor         // 1209600÷4= 302400
	t.maxRetargetTimespan = targetTimespan * adjustmentFactor         // 1209600x4= 4838400
	t.TangleParams = &params
}

type TangleIterator struct {
	Tips       []*msg.Packet
	Relations  *leveldb.DB
	UnApproved *leveldb.DB
	DB         *leveldb.DB
	Err        error
}

var Relations *leveldb.DB = nil
var UnApproved *leveldb.DB = nil
var DataBase *leveldb.DB = nil
var genericLock sync.Mutex

//OpenTangle opens Relations,UnApproved,DataBase
func OpenTangle() (*leveldb.DB, *leveldb.DB, *leveldb.DB, error) {
	if Relations != nil && UnApproved != nil && DataBase != nil {
		return Relations, UnApproved, DataBase, nil
	} else if Relations == nil && UnApproved == nil && DataBase == nil {
		genericLock.Lock()
		defer genericLock.Unlock()
		s, err := storage.OpenFile(Consts.TangleDB, false)
		if err != nil {
			return nil, nil, nil, err
		}
		defer s.Close()
		db, err := leveldb.Open(s, nil)
		if err != nil {
			return nil, nil, nil, err
		}
		DataBase = db

		r, err := storage.OpenFile(Consts.TangleRelations, false)
		if err != nil {
			return nil, nil, nil, err
		}
		defer r.Close()
		re, err := leveldb.Open(r, nil)
		if err != nil {
			return nil, nil, nil, err
		}
		Relations = re

		u, err := storage.OpenFile(Consts.TangleUnApproved, false)
		if err != nil {
			return nil, nil, nil, err
		}
		defer u.Close()
		un, err := leveldb.Open(u, nil)
		if err != nil {
			return nil, nil, nil, err
		}
		UnApproved = un
		return re, un, db, nil
	} else {
		return nil, nil, nil, Consts.ErrInconsistantTangleDB
	}
}

func NewTangle(bc *Blockchain.Blockchain) (*Tangle, error) {
	re, un, db, err := OpenTangle()
	if err != nil {
		return nil, err
	}
	_, err = db.Get([]byte("empty"), nil)
	if err == leveldb.ErrNotFound {
		t := &Tangle{}
		genesis := GenesisBundle(bc)
		//TODO: use add block instead
		raw, _ := proto.Marshal(genesis)
		err = db.Put(
			bytes.Join([][]byte{
				[]byte("b"), genesis.Hash}, []byte{}), raw, nil)
		if err != nil {
			return nil, err
		}
		err = re.Put(genesis.Hash, []byte{}, nil)
		if err != nil {
			return nil, err
		}
		err = un.Put(genesis.Hash, []byte{}, nil)
		if err != nil {
			return nil, err
		}
		//Second Genesis
		genesis = nil
		genesis = GenesisBundle(bc)
		//TODO: use add block instead
		raw, _ = proto.Marshal(genesis)
		err = db.Put(
			bytes.Join([][]byte{
				[]byte("b"), genesis.Hash}, []byte{}), raw, nil)
		if err != nil {
			return nil, err
		}
		err = un.Put(genesis.Hash, []byte{}, nil)
		if err != nil {
			return nil, err
		}
		err = re.Put(genesis.Hash, []byte{}, nil)
		if err != nil {
			return nil, err
		}
		err = db.Put([]byte("empty"), []byte("false"), nil)
		if err != nil {
			return nil, err
		}
		t.DB = db
		t.Relations = re
		t.UnApproved = un
		t.Blockchain = bc
		return t, nil
	}
	return nil, Consts.ErrTangleExists
}

func GenesisBundle(bc *Blockchain.Blockchain) *msg.Packet {
	packet := &msg.Packet{}
	packet.Addr = nil
	packet.CurrentBlockNumber = bc.Tip.CurrentBlockNumber
	packet.Diff = 1
	packet.PacketType = msg.Packet_BUNDLE
	packet.Prev = nil
	packet.Sign = []byte("This is the Tangle genesis")
	packet.Timestamp = ptypes.TimestampNow()
	packet.CurrentBlockHash = bc.Tip.CurrentBlockHash
	bun := &msg.Bundle{}
	bun.Verify1 = nil
	bun.Verify2 = nil
	bun.Verify3 = nil
	bun.Transactions = nil
	bun.Hash = Bundle.GetBundleHash(*bun)
	packet.Data = &msg.Packet_BundleData{bun}
	Packet.SetHash(packet)
	return packet
}

func (t *Tangle) AddBundle(p *msg.Packet, special bool) error {
	t.tangleLock.Lock()
	defer t.tangleLock.Unlock()
	// CRITIAL: no esp check is done here
	// check should be done in other functions
	raw, _ := proto.Marshal(p)
	// 'b' is bundle here
	err := t.DB.Put(
		bytes.Join([][]byte{
			[]byte("b"), p.Hash}, []byte{}), raw, nil)
	if err != nil {
		return err
	}
	rawV1, err := t.Relations.Get(p.GetBundleData().Verify1, nil)
	if err == leveldb.ErrNotFound {
		err = nil
		rawV1 = []byte{}
	}
	if err != nil {
		return err
	}
	rawV2, err := t.Relations.Get(p.GetBundleData().Verify2, nil)
	if err == leveldb.ErrNotFound {
		err = nil
		rawV1 = []byte{}
	}
	if err != nil {
		return err
	}
	rawV1, err = Utils.AppendMarshalSha3(rawV1, p.Hash)
	if err != nil {
		return err
	}
	t.Relations.Put(p.GetBundleData().Verify1, rawV1, nil)
	rawV2, err = Utils.AppendMarshalSha3(rawV2, p.Hash)
	if err != nil {
		return err
	}
	t.Relations.Put(p.GetBundleData().Verify2, rawV1, nil)
	if special {
		rawV3, err := t.Relations.Get(p.GetBundleData().Verify3, nil)
		if err != nil {
			return err
		}
		rawV3, err = Utils.AppendMarshalSha3(rawV3, p.Hash)
		if err != nil {
			return err
		}
		t.Relations.Put(p.GetBundleData().Verify1, rawV3, nil)
	}
	t.UnApproved.Put(p.Hash, []byte{}, nil)
	return nil
}

//This is a temporary implementation of function
func (t *Tangle) PickUnapproved(sep bool) ([]byte, []byte, []byte) {
	//TODO: must use age check and do MCMC
	iter := t.UnApproved.NewIterator(nil, nil)
	iter.Next()
	v1 := make([]byte, 64)
	v2 := make([]byte, 64)
	copy(v1, iter.Key())
	iter.Next()
	copy(v2, iter.Key())
	if sep {
		iter.Next()
		v3 := iter.Key()
		t.UnApproved.Delete(v3, nil)
		return v1, v2, v3
	}
	iter.Release()
	t.UnApproved.Delete(v1, nil)
	return v1, v2, nil
}

//Prev returns the prev Tips of the TangleIterator
func (ti *TangleIterator) Prev() bool {
	var tmpTips []*msg.Packet
	for _, v := range ti.Tips {
		// it must have v1 if it's in DB
		if v.GetBundleData().Verify1 == nil {
			continue
		}
		t, _ := GetPacketFromTangle(v.GetBundleData().Verify1)
		tmpTips = append(tmpTips, t)
		t, _ = GetPacketFromTangle(v.GetBundleData().Verify2)
		tmpTips = append(tmpTips, t)
		if v.GetBundleData().Verify3 != nil {
			t, _ = GetPacketFromTangle(v.GetBundleData().Verify3)
			tmpTips = append(tmpTips, t)
		}
	}
	if len(tmpTips) == 0 {
		return false
	}
	ti.Tips = tmpTips
	return true
}

// Value retrive the current value for iterator
func (ti *TangleIterator) Value() ([]*msg.Packet, error) {
	if ti.Err != nil {
		return nil, ti.Err
	}
	return ti.Tips, nil
}

// ResetErr rests error value for iterator
func (ti *TangleIterator) ResetErr() {
	ti.Err = nil
}

// InitIter initalize the iterator for a Tangle
func (ti *TangleIterator) InitIter(tip []*msg.Packet) error {
	ti.Relations, ti.UnApproved, ti.DB, _ = OpenTangle()
	ti.Tips = tip
	return nil
}

func GetPacketFromTangle(hash []byte) (*msg.Packet, error) {
	// Doesn't support multi Tangle
	raw, err := DataBase.Get(bytes.Join([][]byte{
		[]byte("b"), hash}, []byte{}), nil)
	if err != nil {
		return nil, err
	}
	value := &msg.Packet{}
	proto.Unmarshal(raw, value)
	return value, nil
}

//ExportToJSON exports Tangle as JSON
func (ti *Tangle) ExportToJSON(path string, tips []*msg.Packet) error {
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
	iter := &TangleIterator{}
	err = iter.InitIter(tips)
	if err != nil {
		return err
	}
	v, err := iter.Value()
	if err != nil {
		return err
	}
	for _, i := range v {
		err = m.Marshal(buf, i)
		if err != nil {
			return err
		}
		buf.WriteString(",")
	}
	for iter.Prev() {
		if err != nil {
			return err
		}
		v, err := iter.Value()
		if err != nil {
			return err
		}
		for _, i := range v {
			err = m.Marshal(buf, i)
			if err != nil {
				return err
			}
			buf.WriteString(",")
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

func (ti *Tangle) ReadBundle(hash []byte) (*msg.Packet, error) {
	db := ti.DB
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
	case *msg.Packet_BundleData:
		return block, nil
	default:
		panic(Consts.ErrNotABlock)
	}
}

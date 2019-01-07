package Tangle

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"sync"
	"time"

	"github.com/amidmm/MyChain/Transaction"
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
	UsersTips           *leveldb.DB
	Blockchain          *Blockchain.Blockchain
	InitialCounter      uint64
	CurrentDiff         uint32
	tangleLock          sync.Mutex
	minRetargetTimespan int64  // target timespan / adjustment factor
	maxRetargetTimespan int64  // target timespan * adjustment factor
	blocksPerRetarget   uint64 // target timespan / target time per block
	VblocksPerRetarget  uint64 // same as blocksPerRetarget but for virtual blockchains
	VbcDiffChangeRate   uint64 // Rate of change in diff of virtual blockchain
	TangleParams        *tanleParams
}

type tanleParams struct {
	TargetTimespan           time.Duration
	TargetTimePerBlock       time.Duration
	LastBlockToVerify        uint64 // last block which is acceptable
	RetargetAdjustmentFactor int64
	ReduceMinDifficulty      bool
	MinDiffReductionTime     time.Duration
	GenerateSupported        bool
	MaxBlockDiff             uint64
}

func (t *Tangle) InitTangle() {
	params := tanleParams{
		LastBlockToVerify:        6,
		TargetTimespan:           time.Hour * 24 * 14, // 14 days
		TargetTimePerBlock:       time.Second * 10,    // 10 Seconds
		RetargetAdjustmentFactor: 2,                   // 25% less, 400% more
		ReduceMinDifficulty:      false,
		MinDiffReductionTime:     0,
		GenerateSupported:        false,
	}
	targetTimespan := int64(params.TargetTimespan / time.Second)         //TargetTimespan in seconds
	targetTimePerBlock := int64(params.TargetTimePerBlock / time.Second) //TargetTimePerBlock in seconds
	adjustmentFactor := params.RetargetAdjustmentFactor
	t.blocksPerRetarget = uint64(targetTimespan / targetTimePerBlock) //120960
	t.minRetargetTimespan = targetTimespan / adjustmentFactor         // 1209600÷4= 302400
	t.maxRetargetTimespan = targetTimespan * adjustmentFactor         // 1209600x4= 4838400
	t.VblocksPerRetarget = 100
	t.VbcDiffChangeRate = 2
	t.TangleParams = &params
	t.CurrentDiff = Consts.TangleInitPoWLimit
}

type Iterator struct {
	Tips       []*msg.Packet
	Relations  *leveldb.DB
	UnApproved *leveldb.DB
	DB         *leveldb.DB
	UsersTips  *leveldb.DB
	Err        error
}

var Relations *leveldb.DB = nil
var UnApproved *leveldb.DB = nil
var DataBase *leveldb.DB = nil
var UserTips *leveldb.DB = nil
var genericLock sync.Mutex
var FirstInital msg.Packet
var LastInital msg.Packet

//OpenTangle opens Relations,UnApproved,DataBase
func OpenTangle() (*leveldb.DB, *leveldb.DB, *leveldb.DB, *leveldb.DB, error) {
	if Relations != nil && UnApproved != nil && DataBase != nil && UserTips != nil {
		return Relations, UnApproved, DataBase, UserTips, nil
	} else if Relations == nil && UnApproved == nil && DataBase == nil && UserTips == nil {
		genericLock.Lock()
		defer genericLock.Unlock()
		s, err := storage.OpenFile(Consts.TangleDB, false)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		db, err := leveldb.Open(s, nil)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		DataBase = db

		r, err := storage.OpenFile(Consts.TangleRelations, false)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		re, err := leveldb.Open(r, nil)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		Relations = re

		u, err := storage.OpenFile(Consts.TangleUnApproved, false)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		un, err := leveldb.Open(u, nil)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		UnApproved = un

		uTips, err := storage.OpenFile(Consts.UserTips, false)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		userTips, err := leveldb.Open(uTips, nil)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		UserTips = userTips
		return re, un, db, userTips, nil
	} else {
		return nil, nil, nil, nil, Consts.ErrInconsistantTangleDB
	}
}

func NewTangle(bc *Blockchain.Blockchain) (*Tangle, error) {
	re, un, db, users, err := OpenTangle()
	if err != nil {
		return nil, err
	}
	_, err = db.Get([]byte("empty"), nil)
	if err == leveldb.ErrNotFound {
		t := &Tangle{}
		genesis := GenesisBundle(bc)
		FirstInital = *genesis
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
		LastInital = *genesis
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
		t.UsersTips = users
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
	var data struct {
		Verify1 []byte
		Verify2 []byte
		Verify3 []byte
	}
	switch p.Data.(type) {
	case *msg.Packet_BundleData:
		data.Verify1 = p.GetBundleData().Verify1
		data.Verify2 = p.GetBundleData().Verify2
		data.Verify2 = p.GetBundleData().Verify2
	case *msg.Packet_InitialData:
		data.Verify1 = p.GetInitialData().Verify1
		data.Verify2 = p.GetInitialData().Verify2
		data.Verify3 = nil
	}
	rawV1, err := t.Relations.Get(data.Verify1, nil)
	if err == leveldb.ErrNotFound {
		err = nil
		rawV1 = []byte{}
	}
	if err != nil {
		return err
	}
	rawV2, err := t.Relations.Get(data.Verify2, nil)
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
	t.Relations.Put(data.Verify1, rawV1, nil)
	rawV2, err = Utils.AppendMarshalSha3(rawV2, p.Hash)
	if err != nil {
		return err
	}
	t.Relations.Put(data.Verify2, rawV1, nil)
	if special && p.PacketType == msg.Packet_BUNDLE {
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
	t.UsersTips.Put(p.Addr, p.Hash, nil)
	rawCounter, err := t.UsersTips.Get(bytes.Join([][]byte{
		[]byte("c"), p.Addr}, []byte{}), nil)
	if err == leveldb.ErrNotFound {
		t.UsersTips.Put(bytes.Join([][]byte{
			[]byte("c"), p.Addr}, []byte{}), t.InitialPacketCounter(p), nil)
	} else {
		counter := Utils.UnMarshalPacketCounter(rawCounter)
		counter[p.CurrentBlockNumber]++
		rawCounter, err = Utils.MarshalPacketCounter(counter, t.TangleParams.LastBlockToVerify)
		if err != nil {
			return err
		}
		t.UsersTips.Put(bytes.Join([][]byte{
			[]byte("c"), p.Addr}, []byte{}), rawCounter, nil)
	}

	bufBlockNumber := make([]byte, 8)
	binary.PutUvarint(bufBlockNumber, p.CurrentBlockNumber)
	t.UnApproved.Put(p.Hash, bufBlockNumber, nil)
	if p.PacketType == msg.Packet_INITIAL {
		t.InitialCounter++
		LastInital = *p
		if t.InitialCounter >= t.blocksPerRetarget {
			diff, err := t.CalcNextRequiredDifficulty()
			if err != nil {
				return err
			}
			t.CurrentDiff = diff
		}
	}
	if p.PacketType == msg.Packet_BUNDLE {
		if ok, err := Transaction.HandleBundle(p); err != nil || !ok {
			return err
		}
	}
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

//Prev returns the prev Tips of the Iterator
func (ti *Iterator) Prev() bool {
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

//Next returns the next Tips of the Iterator
func (ti *Iterator) Next() bool {
	var tmpTips []*msg.Packet
	for _, tip := range ti.Tips {
		rawOut, err := ti.Relations.Get(tip.Hash, nil)
		if err == leveldb.ErrNotFound {
			continue
		}
		if err != nil {
			return false
		}
		sha3Out, err := Utils.UnMarshalSha3List(rawOut)
		if err != nil {
			return false
		}
		for _, v := range sha3Out {
			p, _ := GetPacketFromTangle(v)
			tmpTips = append(tmpTips, p)
		}
	}
	if len(tmpTips) == 0 {
		return false
	}
	ti.Tips = tmpTips
	return true
}

// Value retrive the current value for iterator
func (ti *Iterator) Value() ([]*msg.Packet, error) {
	if ti.Err != nil {
		return nil, ti.Err
	}
	return ti.Tips, nil
}

// ResetErr rests error value for iterator
func (ti *Iterator) ResetErr() {
	ti.Err = nil
}

// InitIter initalize the iterator for a Tangle
func (ti *Iterator) InitIter(tip []*msg.Packet) error {
	ti.Relations, ti.UnApproved, ti.DB, ti.UsersTips, _ = OpenTangle()
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
func (t *Tangle) ExportToJSON(path string, tips []*msg.Packet, reverse bool) error {
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
	iter := &Iterator{}
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
	var directionFunc func() bool
	directionFunc = iter.Prev
	if reverse == true {
		directionFunc = iter.Next
	}
	for directionFunc() {
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

func (t *Tangle) ReadBundle(hash []byte) (*msg.Packet, error) {
	db := t.DB
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

func (t *Tangle) RelativeAncestor(p *msg.Packet, distance uint64) ([]*msg.Packet, error) {
	if distance == 0 {
		return nil, Consts.ErrWrongParam
	}
	tips := []*msg.Packet{}
	tips = append(tips, p)
	iter := &Iterator{}
	iter.InitIter(tips)
	exits := iter.Prev()
	for exits && distance > 1 {
		distance--
		exits = iter.Prev()
	}
	if !exits && distance > 1 {
		return nil, Consts.ErrWrongParam
	}
	return iter.Value()
}

func (t *Tangle) Close() error {
	if (t.DB == nil || DataBase == nil) &&
		(t.Relations == nil || Relations == nil) && (t.UnApproved == nil || UnApproved == nil) &&
		(t.UsersTips == nil || UserTips == nil) {
		return leveldb.ErrClosed
	}
	t.tangleLock.Lock()
	defer t.tangleLock.Unlock()
	DataBase = nil
	Relations = nil
	UnApproved = nil
	UserTips = nil
	t.DB.Close()
	t.Relations.Close()
	t.UnApproved.Close()
	t.UsersTips.Close()
	log.Println("Tangle databases is down...")
	return nil
}

func (t *Tangle) CalcNextRequiredDifficulty() (uint32, error) {
	if (t.InitialCounter+1)%t.NextRetarget() != 0 {
		return t.CurrentDiff, nil
	}
	firstNode := FirstInital
	actualTimespan := LastInital.Timestamp.Seconds - firstNode.Timestamp.Seconds
	adjustedTimespan := actualTimespan
	if actualTimespan < t.minRetargetTimespan {
		adjustedTimespan = t.minRetargetTimespan
	} else if actualTimespan > t.maxRetargetTimespan {
		adjustedTimespan = t.maxRetargetTimespan
	}
	targetTimeSpan := int64(t.TangleParams.TargetTimespan / time.Second)
	diff := int64(LastInital.Diff) * (targetTimeSpan / adjustedTimespan)
	if uint32(diff) < Consts.TangleInitPoWLimit {
		return Consts.TangleInitPoWLimit, nil
	}
	FirstInital = LastInital
	return uint32(diff), nil
}

func (t *Tangle) NextRetarget() uint64 {
	retarget := t.blocksPerRetarget - (t.InitialCounter % t.blocksPerRetarget)
	return retarget + t.InitialCounter
}

func (t *Tangle) InitialPacketCounter(p *msg.Packet) []byte {
	data := make(map[uint64]uint64)
	data[p.CurrentBlockNumber]++
	raw, err := Utils.MarshalPacketCounter(data, t.TangleParams.LastBlockToVerify)
	if err != nil {
		return nil
	}
	return raw
}

func (t *Tangle) NextVbcRetarget(addr []byte) uint64 {
	// TODO: more sophisticated method should be used
	sum, _ := t.CurrentVbcCounter(addr)
	retarget := t.VblocksPerRetarget - (sum % t.VblocksPerRetarget)
	return retarget + sum
}

// CurrentVbcCounter reads current number for virtual blockchain
func (t *Tangle) CurrentVbcCounter(addr []byte) (uint64, error) {
	rawCounter, err := t.UsersTips.Get(bytes.Join([][]byte{
		[]byte("c"), addr}, []byte{}), nil)
	if err == leveldb.ErrNotFound {
		return 1, nil
	}
	sum := uint64(0)
	counter := Utils.UnMarshalPacketCounter(rawCounter)
	for _, v := range counter {
		sum += v
	}
	return sum, nil
}

func (t *Tangle) CalcNextVbcDiffProducer(addr []byte) (uint32, error) {
	c, err := t.CurrentVbcCounter(addr)
	if err != nil {
		return 0, err
	}
	c++
	diff := c / t.VblocksPerRetarget
	diff *= t.VbcDiffChangeRate
	if uint32(diff) < Consts.VbcPoWLimit {
		return Consts.VbcPoWLimit, nil
	}
	return uint32(diff), nil
}

func (t *Tangle) CalcNextVbcDiff(addr []byte) (uint32, error) {
	c, err := t.CurrentVbcCounter(addr)
	if err != nil {
		return 0, err
	}
	diff := c / t.VblocksPerRetarget
	diff *= t.VbcDiffChangeRate
	if uint32(diff) < Consts.VbcPoWLimit {
		return Consts.VbcPoWLimit, nil
	}
	return uint32(diff), nil
}

func (t *Tangle) HasSeenInTangle(p *msg.Packet) bool {
	db := t.DB
	_, err := db.Get(bytes.Join(
		[][]byte{[]byte("b"), p.Hash}, []byte{}),
		nil)
	if err == nil {
		return true
	}
	return false
}

func (t *Tangle) ExportVBCToJSON(path string, addr []byte) error {
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
	raw, err := t.UsersTips.Get(addr, nil)
	if err != nil {
		return err
	}
	p := &msg.Packet{}
	proto.Unmarshal(raw, p)
	m.Marshal(buf, p)
	for p.Prev == nil {
		_, err = buf.WriteString(",")
		if err != nil {
			return err
		}
		raw, err = t.DB.Get(bytes.Join(
			[][]byte{[]byte("b"), p.Prev}, []byte{}),
			nil)
		p = &msg.Packet{}
		err = proto.Unmarshal(raw, p)
		m.Marshal(buf, p)
	}
	buf.WriteString("]")
	if err != nil {
		return err
	}
	_, err = f.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

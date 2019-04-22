package Miner

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/amidmm/MyChain/Account"

	"github.com/amidmm/MyChain/Generator"

	"github.com/amidmm/MyChain/Blockchain"

	"github.com/amidmm/MyChain/Messages"
	"github.com/amidmm/MyChain/Packet"
	"github.com/amidmm/MyChain/PoW"
	"github.com/golang/protobuf/ptypes"
	crypto "github.com/libp2p/go-libp2p-crypto"
)

var CurrentWeak []*msg.WeakReq
var CurrentDumbPacketHashes []*msg.HashArray

// Busy the it is 1
var IsBusy int32 = 0
var CtxNewBlock context.Context
var CtxNewBlockCancel context.CancelFunc

func init() {
	CtxNewBlock = context.Background()
	CtxNewBlock, CtxNewBlockCancel = context.WithCancel(CtxNewBlock)
	CurrentWeak = []*msg.WeakReq{}
	CurrentDumbPacketHashes = []*msg.HashArray{}
}

func ShouldBuildBlock(bc *Blockchain.Blockchain) bool {
	//Here the algorithm calculate the payoff
	//for now we jsut use a dumb strategy
	if len(CurrentDumbPacketHashes)+len(CurrentWeak) > 100 {
		return true
	}
	if bc.Tip.Timestamp == nil {
		return true
	}
	//if changed then change the miner Sync ticker too
	if time.Now().Unix()-bc.Tip.Timestamp.Seconds > 10 {
		return true
	}
	return false
}

func BuildBlock(bc *Blockchain.Blockchain, u *Account.User) msg.Packet {
	atomic.SwapInt32(&IsBusy, 1)
	defer func() {
		atomic.SwapInt32(&IsBusy, 0)
	}()
	CtxNewBlock = context.Background()
	CtxNewBlock, CtxNewBlockCancel = context.WithCancel(CtxNewBlock)
	diff, _ := bc.CalcNextRequiredDifficulty()
	PrevHash := bc.Tip.Hash
	packet := &msg.Packet{}
	packet.Addr, _ = crypto.MarshalPublicKey(u.PubKey)
	packet.CurrentBlockNumber = bc.Tip.CurrentBlockNumber + 1
	packet.Diff = diff
	packet.PacketType = msg.Packet_BLOCK
	packet.Prev = PrevHash
	packet.Timestamp = ptypes.TimestampNow()
	block := &msg.Block{}
	block.Reqs = CurrentWeak
	//change sanity
	block.Sanities = []*msg.SanityCheck{}
	block.PacketHashs = CurrentDumbPacketHashes
	valueBase, _ := bc.ExpectedBlockReward()
	block.Coinbase = Generator.GenInTx(u, valueBase, true)
	packet.Data = &msg.Packet_BlockData{block}
	Packet.SetPacketSign(packet, u)
	PoW.SetPoW(CtxNewBlock, packet, diff)
	if packet.Nonce == nil {
		CurrentWeak = make([]*msg.WeakReq, 10)
		CurrentDumbPacketHashes = make([]*msg.HashArray, 10)
		return msg.Packet{}
	}
	Packet.SetHash(packet)
	return *packet
}

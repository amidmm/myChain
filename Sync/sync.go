package Sync

import (
	"time"

	"github.com/amidmm/MyChain/Packet"

	"github.com/amidmm/MyChain/PoW"

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

//Prev Should be checked in the packet type specific methods
func IncomingPacketValidation(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) (bool, error) {
	var packetType msg.PacketType = -1
	var diffFunc func() (uint32, error)
	switch p.Data.(type) {
	case *msg.Packet_BlockData:
		packetType = msg.Packet_BLOCK
		diffFunc = bc.CalcNextRequiredDifficulty
	case *msg.Packet_BundleData:
		packetType = msg.Packet_BUNDLE
	case *msg.Packet_RepData:
		packetType = msg.Packet_REP
	case *msg.Packet_WeakData:
		packetType = msg.Packet_WEAKREQ
	case *msg.Packet_SanityData:
		packetType = msg.Packet_SANITYCHECK
	case *msg.Packet_InitialData:
		packetType = msg.Packet_INITIAL
		diffFunc = t.CalcNextRequiredDifficulty
	}
	if packetType != p.PacketType {
		return false, nil
	}
	if diffFunc == nil {
		diff, err := t.CalcNextVbcDiff(p.Addr)
		if err != nil {
			return false, err
		}
		if p.Diff < diff {
			return false, nil
		}
	} else {
		diff, err := diffFunc()
		if err != nil {
			return false, err
		}
		if p.Diff < diff {
			return false, nil
		}
	}
	if r, err := PoW.ValidatePoW(*p, p.Diff); err != nil || !r {
		return false, err
	}
	//TODO: is it needed???
	lim := time.Now().Unix() - int64(Consts.TimestampBound.Seconds())
	if p.Timestamp.Seconds < lim {
		return false, nil
	}

	if p.PacketType != msg.Packet_BLOCK {
		if int64(p.CurrentBlockNumber) < int64(bc.Tip.CurrentBlockNumber-t.TangleParams.LastBlockToVerify) {
			return false, nil
		}
		if block, err := bc.ReadBlock(p.CurrentBlockHash); err != nil || block.CurrentBlockNumber != p.CurrentBlockNumber {
			return false, err
		}
	}
	if ok, err := Packet.ValidateSign(p); err != nil || !ok {
		return false, err
	}
	return true, nil
}

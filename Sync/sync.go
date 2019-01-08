package Sync

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/amidmm/MyChain/Packet"
	"github.com/amidmm/MyChain/Validator"

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

func IncomingPacket(ctx context.Context, packetChan <-chan *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) {
	for {
		select {
		case p := <-packetChan:
			hash := sha1.New()
			hash.Write(p.Hash)
			if ok, err := IncomingPacketValidation(p, bc, t); err != nil || !ok {
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				log.Println("\033[31m Sync: Invalid packet:\t" + hex.EncodeToString(hash.Sum(nil)) + "\t" + errStr + "\033[0m")
				continue
			}
			ok, err := CheckOrdered(p, bc, t)
			if err != nil {
				log.Println("\033[31m Sync: Invalid packet:\t" + hex.EncodeToString(hash.Sum(nil)) + "\t" + err.Error() + "\033[0m")
			}
			if !ok {
				if _, err = t.UsersTips.Get(p.Addr, nil); p.PacketType != msg.Packet_INITIAL && p.PacketType != msg.Packet_BLOCK && err != nil {
					log.Println("\033[31m Sync: DOS protection:\t" + hex.EncodeToString(hash.Sum(nil)) + "\t" + err.Error() + "\033[0m")
				}
				raw, _ := proto.Marshal(p)
				err = UnsyncPoolDB.Put(p.Prev, raw, nil)
				log.Println("\033[34m Sync: added unordered packet:\t" + hex.EncodeToString(hash.Sum(nil)) + "\033[0m")
				continue
			}
			unsyncRaw, err := UnsyncPoolDB.Get(p.Hash, nil)
			if err == leveldb.ErrNotFound {
				if ok, err := ProcessPacket(p, bc, t); err != nil || !ok {
					errStr := ""
					if err != nil {
						errStr = err.Error()
					}
					log.Println("\033[31m Sync: unable to process packet:\t" + hex.EncodeToString(hash.Sum(nil)) + "\t" + errStr + "\033[0m")
					continue
				}
				log.Println("\033[32m Sync: Packet processed:\t" + hex.EncodeToString(hash.Sum(nil)) + "\033[0m")
			} else {
				unsyncPacket := &msg.Packet{}
				proto.Unmarshal(unsyncRaw, unsyncPacket)
				unsyncSha1 := sha1.New()
				unsyncSha1.Write(unsyncPacket.Hash)
				if ok, err := ProcessPacket(unsyncPacket, bc, t); err != nil || !ok {
					errStr := ""
					if err != nil {
						errStr = err.Error()
					}
					log.Println("\033[31m Sync: unable to process packet:\t" + hex.EncodeToString(unsyncSha1.Sum(nil)) + "\t" + errStr + "\033[0m")
					UnsyncPoolDB.Delete(unsyncPacket.Hash, nil)
					continue
				}
				log.Println("\033[32m Sync: Packet processed:\t" + hex.EncodeToString(unsyncSha1.Sum(nil)) + "\033[0m")
				if ok, err := ProcessPacket(p, bc, t); err != nil || !ok {
					errStr := ""
					if err != nil {
						errStr = err.Error()
					}
					log.Println("\033[31m Sync: unable to process packet:\t" + hex.EncodeToString(hash.Sum(nil)) + "\t" + errStr + "\033[0m")
					continue
				}
				log.Println("\033[32m Sync: Packet processed:\t" + hex.EncodeToString(hash.Sum(nil)) + "\033[0m")
			}
		case <-ctx.Done():
			log.Println("\033[41m Sync: exiting \033[0m")
			return
		}
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

func CheckOrdered(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) (bool, error) {
	var err error
	joined := bytes.Join([][]byte{
		[]byte("b"), p.Prev}, []byte{})
	if p.PacketType == msg.Packet_BLOCK {
		_, err = bc.DB.Get(joined, nil)
	} else if p.PacketType != msg.Packet_INITIAL {
		_, err = t.DB.Get(joined, nil)
	}
	if err == leveldb.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func ProcessPacket(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) (bool, error) {
	return Validator.Validate(p, bc, t)
	//return false, Consts.ErrNotImplemented
}

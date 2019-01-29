package Sync

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/amidmm/MyChain/Transaction"

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
var SyncMode = false
var syncLock sync.Mutex
var UnsyncPoolDBTx *leveldb.DB

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

func IncomingPacket(ctx context.Context, packetChan <-chan *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle, advertiserChan chan *msg.Packet) {
	for {
		select {
		case x := <-packetChan:
			p := x
			hash := sha1.New()
			hash.Write(p.Hash)
			if HasBeenSeen(p, bc, t) {
				log.Println("\033[31m Sync: repetead packet\033[0m")
				continue
			}
			if ok, err := IncomingPacketValidation(p, bc, t); err != nil || !ok {
				hash := sha1.New()
				hash.Write(p.Hash)
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				log.Println("\033[31m Sync: Invalid packet1:\t" + hex.EncodeToString(hash.Sum(nil)) + "\t" + errStr + "\033[0m")
				continue
			}
			if ok := PreProcess(p, bc, t, advertiserChan); !ok {
				continue
			}
			Crawler(p, bc, t, advertiserChan)

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

	//An easy way to prevent DOS attacks
	if !SyncMode {
		lim := time.Now().Unix() - int64(Consts.TimestampBound.Seconds())
		if p.Timestamp.Seconds < lim {
			return false, nil
		}
	}

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
	if p.PacketType != msg.Packet_BLOCK && SyncMode == false {
		if int64(p.CurrentBlockNumber) < int64(bc.Tip.CurrentBlockNumber-t.TangleParams.LastBlockToVerify) {
			return false, nil
		}
	}
	if ok, err := Packet.ValidateSign(p); err != nil || !ok {
		return false, err
	}
	return true, nil
}

func CheckOrdered(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) (bool, [][]byte, error) {
	var err error
	var listOfUnordered [][]byte
	isOrderd := true
	var data struct {
		Verify1 []byte
		Verify2 []byte
		Verify3 []byte
	}
	switch p.Data.(type) {
	case *msg.Packet_BundleData:
		data.Verify1 = p.GetBundleData().Verify1
		data.Verify2 = p.GetBundleData().Verify2
		data.Verify3 = p.GetBundleData().Verify3
		if p.GetBundleData().Transactions != nil {
			for _, v := range p.GetBundleData().Transactions {
				if v.Value < 0 {
					_, err := Transaction.GetUTXO(v.RefTx)
					if err == leveldb.ErrNotFound {
						listOfUnordered = append(listOfUnordered, v.RefTx)
						isOrderd = false
					}
				}
			}
		}
	case *msg.Packet_InitialData:
		data.Verify1 = p.GetInitialData().Verify1
		data.Verify2 = p.GetInitialData().Verify2
		data.Verify3 = nil
		if p.GetInitialData().PoBurn != nil {
			for _, v := range p.GetInitialData().PoBurn.Transactions {
				if v.Value < 0 {
					_, err := Transaction.GetUTXO(v.RefTx)
					if err == leveldb.ErrNotFound {
						listOfUnordered = append(listOfUnordered, v.RefTx)
						isOrderd = false
					}
				}
			}
		}
	case *msg.Packet_RepData:
		data.Verify1 = p.GetRepData().Verify1
		data.Verify2 = p.GetRepData().Verify2
		data.Verify3 = nil
	case *msg.Packet_WeakData:
		data.Verify1 = p.GetWeakData().Verify1
		data.Verify2 = p.GetWeakData().Verify2
		data.Verify3 = nil
		if p.GetWeakData().Burn != nil {
			for _, v := range p.GetWeakData().Burn.Transactions {
				if v.Value < 0 {
					_, err := Transaction.GetUTXO(v.RefTx)
					if err == leveldb.ErrNotFound {
						listOfUnordered = append(listOfUnordered, v.RefTx)
						isOrderd = false
					}
				}
			}
		}
	case *msg.Packet_SanityData:
		data.Verify1 = p.GetSanityData().Verify1
		data.Verify2 = p.GetSanityData().Verify2
		data.Verify3 = nil
	}

	joined := bytes.Join([][]byte{
		[]byte("b"), p.Prev}, []byte{})

	if p.PacketType == msg.Packet_BLOCK {
		_, err = bc.DB.Get(joined, nil)
		if err == leveldb.ErrNotFound {
			listOfUnordered = append(listOfUnordered, p.Prev)
			isOrderd = false
		}
	} else if p.PacketType != msg.Packet_INITIAL {
		_, err = t.DB.Get(joined, nil)
		if err == leveldb.ErrNotFound {
			listOfUnordered = append(listOfUnordered, p.Prev)
			isOrderd = false
		}
	}

	if p.PacketType != msg.Packet_BLOCK {
		joined = bytes.Join([][]byte{
			[]byte("b"), data.Verify1}, []byte{})
		_, err1 := t.DB.Get(joined, nil)
		if err1 == leveldb.ErrNotFound {
			listOfUnordered = append(listOfUnordered, data.Verify1)
			isOrderd = false
		}
		joined = bytes.Join([][]byte{
			[]byte("b"), data.Verify2}, []byte{})
		_, err2 := t.DB.Get(joined, nil)
		if err2 == leveldb.ErrNotFound {
			listOfUnordered = append(listOfUnordered, data.Verify2)
			isOrderd = false
		}
		if data.Verify3 != nil {
			joined = bytes.Join([][]byte{
				[]byte("b"), data.Verify3}, []byte{})
			_, err3 := t.DB.Get(joined, nil)
			if err3 == leveldb.ErrNotFound {
				listOfUnordered = append(listOfUnordered, data.Verify3)
				isOrderd = false
			}
		}
		joined = bytes.Join([][]byte{
			[]byte("b"), p.CurrentBlockHash}, []byte{})
		_, err1 = bc.DB.Get(joined, nil)
		if err1 == leveldb.ErrNotFound {
			listOfUnordered = append(listOfUnordered, p.CurrentBlockHash)
			isOrderd = false
		}
	}
	return isOrderd, listOfUnordered, err
}

func ProcessPacket(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) (bool, error) {
	return Validator.Validate(p, bc, t)
	//return false, Consts.ErrNotImplemented
}

func ExtractLinks(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) ([][]byte, error) {
	var err error
	var listLinks [][]byte
	switch p.Data.(type) {
	case *msg.Packet_BundleData:
		if p.GetBundleData().Transactions != nil {
			for _, v := range p.GetBundleData().Transactions {
				if v.Value > 0 {
					listLinks = append(listLinks, v.Hash)
				}
			}
		}
	case *msg.Packet_InitialData:
		if p.GetInitialData().PoBurn != nil {
			for _, v := range p.GetInitialData().PoBurn.Transactions {
				if v.Value > 0 {
					listLinks = append(listLinks, v.Hash)
				}
			}
		}
	case *msg.Packet_WeakData:
		if p.GetWeakData().Burn != nil {
			for _, v := range p.GetWeakData().Burn.Transactions {
				if v.Value > 0 {
					listLinks = append(listLinks, v.Hash)
				}
			}
		}
	}
	if p.PacketType == msg.Packet_BLOCK {
		listLinks = append(listLinks, p.GetBlockData().Coinbase.Hash)
	}
	listLinks = append(listLinks, p.Hash)
	return listLinks, err
}

func PreProcess(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle, advertiserChan chan *msg.Packet) bool {
	var err error
	hash := sha1.New()
	hash.Write(p.Hash)
	ok, list, _ := CheckOrdered(p, bc, t)
	if !ok {
		if SyncMode == false {
			if _, err = t.UsersTips.Get(p.Addr, nil); p.PacketType != msg.Packet_INITIAL && p.PacketType != msg.Packet_BLOCK && err != nil {
				log.Println("\033[31m Sync: DOS protection:\t" + hex.EncodeToString(hash.Sum(nil)) + "\t" + err.Error() + "\033[0m")
				return false
			}
		}
		raw, _ := proto.Marshal(p)
		for index := 0; index < len(list); index++ {
			err = UnsyncPoolDB.Put(bytes.Join([][]byte{list[index], p.Hash}, []byte{}), raw, nil)
			_ = err
		}
		log.Println("\033[34m Sync: added unordered packet:\t" + hex.EncodeToString(hash.Sum(nil)) + "\033[0m")
		return false
	}
	if ok, err := ProcessPacket(p, bc, t); err != nil || !ok {
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		log.Println("\033[31m Sync: unable to process packet:\t" + hex.EncodeToString(hash.Sum(nil)) + "\t" + errStr + "\033[0m")
		return false
	}
	log.Println("\033[32m Sync: Packet processed:\t" + hex.EncodeToString(hash.Sum(nil)) + "\033[0m")
	advertiserChan <- p
	return true
}

func Crawler(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle, advertiserChan chan *msg.Packet) {
	list, _ := ExtractLinks(p, bc, t)
	for index := 0; index < len(list); index++ {
		start := bytes.Join([][]byte{list[index], Consts.Empty}, []byte{})
		limit := bytes.Join([][]byte{list[index], Consts.Full}, []byte{})
		iter := UnsyncPoolDB.NewIterator(&util.Range{Start: start, Limit: limit}, nil)
		for iter.Next() {
			prevPacket := &msg.Packet{}
			proto.Unmarshal(iter.Value(), prevPacket)
			if prevPacket.Hash != nil {
				if ok := PreProcess(prevPacket, bc, t, advertiserChan); ok {
					UnsyncPoolDB.Delete(iter.Key(), nil)
					Crawler(prevPacket, bc, t, advertiserChan)
				}
			}
		}
	}
}

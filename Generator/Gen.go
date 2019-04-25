package Generator

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/amidmm/MyChain/Weak"

	"github.com/multiformats/go-multiaddr"

	"github.com/amidmm/MyChain/Consts"

	"github.com/libp2p/go-libp2p-crypto"

	"github.com/amidmm/MyChain/Packet"

	"github.com/amidmm/MyChain/Account"

	"github.com/amidmm/MyChain/Bundle"

	"github.com/amidmm/MyChain/Transaction"

	"github.com/amidmm/MyChain/Blockchain"
	"github.com/amidmm/MyChain/Messages"
	"github.com/amidmm/MyChain/PoW"
	"github.com/golang/protobuf/ptypes"
)

var initailTX []msg.Tx
var count = 0

// GenMultiAccountBlockchain generates number of blocks with multiple users
func GenMultiAccountBlockchain(bc *Blockchain.Blockchain, users []*Account.User, number int, packetChan chan *msg.Packet) {
	var wg sync.WaitGroup
	for _, u := range users {
		wg.Add(1)
		go Gen(bc, u, number, &wg, packetChan)
	}
	wg.Wait()
}

// Gen generates number of blocks
func Gen(bc *Blockchain.Blockchain, u *Account.User, number int, wg *sync.WaitGroup, PacketChan chan *msg.Packet) {
	diff := Consts.PoWLimit
	for number > 0 {
		number--
		PrevHash := bc.Tip.Hash
		packet := &msg.Packet{}
		packet.Addr, _ = crypto.MarshalPublicKey(u.PubKey)
		packet.CurrentBlockNumber = bc.Tip.CurrentBlockNumber + 1
		packet.Diff = diff
		packet.PacketType = msg.Packet_BLOCK
		packet.Prev = PrevHash
		packet.Timestamp = ptypes.TimestampNow()
		block := &msg.Block{}
		block.Reqs = []*msg.Packet{}
		block.Sanities = []*msg.SanityCheck{}
		block.PacketHashs = []*msg.HashArray{}
		valueBase, _ := bc.ExpectedBlockReward()
		block.Coinbase = GenInTx(u, valueBase, true)
		packet.Data = &msg.Packet_BlockData{block}
		Packet.SetPacketSign(packet, u)
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		go func() {
			for {
				time.Sleep(100 * time.Millisecond)
				if packet.CurrentBlockNumber <= bc.Tip.CurrentBlockNumber {
					cancel()
					return
				}
			}
		}()
		PoW.SetPoW(ctx, packet, diff)
		if packet.Nonce == nil {
			continue
		}
		Packet.SetHash(packet)
		PrevHash = packet.Hash
		// _, err := bc.ValidateBlock(packet)
		// fmt.Printf("\033[31m %s \033[0m", err)
		if PacketChan == nil {
			bc.AddBlock(packet)
		} else {
			PacketChan <- packet
		}
		diff, _ = bc.CalcNextRequiredDifficulty()
		hash := sha1.New()
		hash.Write(packet.Hash)
		fmt.Printf("\033[31m %s \t %s \t %d\t\033[0m\n", u.Name, hex.EncodeToString(hash.Sum(nil)), packet.CurrentBlockNumber)
	}
	wg.Done()
}

// GenInTx generates transactions
func GenInTx(u *Account.User, value int64, base bool) *msg.Tx {
	tx := &msg.Tx{}
	rand.Read(tx.Tag)
	tx.Value = value
	if base {
		tx.RefTx = nil
		tx.Hash = Transaction.GetTxHash(*tx)
		tx.Sign, _ = crypto.MarshalPublicKey(u.PubKey)
	} else if value < 0 {
		Foundtx := &msg.Tx{}
		for {
			Foundtx = u.GetUTXO(-value)
			if Foundtx.Hash != nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		tx.RefTx = Foundtx.Hash
		tx.Hash = Transaction.GetTxHash(*tx)
		Transaction.SetTxSign(tx, u)
		count++
		return tx
	} else {
		tx.RefTx = nil
		tx.Hash = Transaction.GetTxHash(*tx)
		tx.Sign, _ = crypto.MarshalPublicKey(u.PubKey)
	}
	return tx
}

// GenBundleWeak generates Bundle without verifications
func GenBundleWeak(u *Account.User) *msg.Bundle {
	bun := &msg.Bundle{}
	bun.BundleType = msg.Bundle_WEAK
	value := rand.Int63n(20)
	tmp := value
	for value > 1 {
		txValue := rand.Int63n(value)
		value -= txValue
		bun.Transactions = append(bun.Transactions, GenInTx(u, txValue, false))
	}
	bun.Transactions = append(bun.Transactions, GenInTx(u, 1, false))
	/////////////////////////////
	value = tmp
	for value > 1 {
		txValue := rand.Int63n(value)
		value -= txValue
		bun.Transactions = append(bun.Transactions, GenInTx(u, -txValue, false))
	}
	bun.Transactions = append(bun.Transactions, GenInTx(u, -1, false))
	bun.Hash = Bundle.GetBundleHash(*bun)
	for _, tx := range bun.Transactions {
		tx.BundleHash = bun.Hash
	}
	return bun
}

// GenInitialTx returns 1000 coinbase-like tx
func GenInitialTx(u *Account.User) {
	Transaction.OpenUTXO()
	x, _ := u.PubKey.Bytes()
	for i := 0; i < 100000; i++ {
		tx := GenInTx(u, 9999999999, true)
		Transaction.PutUTXO(tx, x)
		initailTX = append(initailTX, *tx)
	}
}

func GenEmptyPacket(bc *Blockchain.Blockchain, u *Account.User, msgType msg.PacketType) *msg.Packet {
	PrevHash := bc.Tip.Hash
	packet := &msg.Packet{}
	packet.Addr, _ = crypto.MarshalPublicKey(u.PubKey)
	packet.CurrentBlockNumber = bc.Tip.CurrentBlockNumber
	packet.CurrentBlockHash = bc.Tip.Hash
	packet.Diff = Consts.PoWLimit
	packet.PacketType = msgType
	packet.Prev = PrevHash
	packet.Timestamp = ptypes.TimestampNow()
	return packet
}

// GenInitialPackets generates fake initial packets
func GenInitialPackets(u *Account.User) *msg.Initial {
	inital := &msg.Initial{}
	inital.OwnerAddr, _ = u.PubKey.Bytes()
	// inital.PoBurn = GenBundleWeak(u)
	rand.Read(inital.Service)
	addr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/1234")
	inital.IpfsDetail = addr.Bytes()
	return inital
}

func GenBlock(bc *Blockchain.Blockchain, u *Account.User) *msg.Packet {
	diff, _ := bc.CalcNextRequiredDifficulty()
	if diff < Consts.PoWLimit {
		diff = Consts.PoWLimit
	}
	PrevHash := bc.Tip.Hash
	packet := &msg.Packet{}
	packet.Addr, _ = crypto.MarshalPublicKey(u.PubKey)
	packet.CurrentBlockNumber = bc.Tip.CurrentBlockNumber + 1
	packet.Diff = diff
	packet.PacketType = msg.Packet_BLOCK
	packet.Prev = PrevHash
	packet.Timestamp = ptypes.TimestampNow()
	block := &msg.Block{}
	block.Reqs = []*msg.Packet{}
	block.Sanities = []*msg.SanityCheck{}
	block.PacketHashs = []*msg.HashArray{}
	valueBase, _ := bc.ExpectedBlockReward()
	block.Coinbase = GenInTx(u, valueBase, true)
	packet.Data = &msg.Packet_BlockData{block}
	Packet.SetPacketSign(packet, u)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			if packet.CurrentBlockNumber <= bc.Tip.CurrentBlockNumber {
				cancel()
				return
			}
		}
	}()
	PoW.SetPoW(ctx, packet, diff)
	if packet.Nonce == nil {
		return nil
	}
	Packet.SetHash(packet)
	PrevHash = packet.Hash
	// _, err := bc.ValidateBlock(packet)
	// fmt.Printf("\033[31m %s \033[0m", err)
	fmt.Printf("\033[31m %s \033[0m", u.Name)
	fmt.Println(packet.CurrentBlockNumber)
	return packet
}

func GenWeakReq(u *Account.User, totalTx uint32, totalFee uint32, burn *msg.Bundle) *msg.WeakReq {
	weakreq := &msg.WeakReq{}
	weakreq.TotalTx = totalTx
	weakreq.TotalFee = totalFee
	weakreq.Burn = burn
	weakreq.Hash = Weak.GetWeakReqHash(*weakreq)
	return weakreq
}

func GenBurnTX(u *Account.User, value uint32) *msg.Bundle {
	bun := &msg.Bundle{}
	bun.BundleType = msg.Bundle_WEAK
	bun.Transactions = append(bun.Transactions, GenInTx(u, -int64(value), false))
	bun.Transactions = append(bun.Transactions, GenInBurnTX(u, int64(value), false))
	/////////////////////////////
	bun.Hash = Bundle.GetBundleHash(*bun)
	for _, tx := range bun.Transactions {
		tx.BundleHash = bun.Hash
	}
	return bun
}

func GenInBurnTX(u *Account.User, value int64, base bool) *msg.Tx {

	tx := &msg.Tx{}
	rand.Read(tx.Tag)
	tx.Value = value

	tx.RefTx = nil
	tx.Hash = Transaction.GetTxHash(*tx)
	tx.Sign = Consts.Empty

	return tx

}

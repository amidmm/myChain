package Validator

import (
	"bytes"
	"log"

	"github.com/amidmm/MyChain/Blockchain"
	"github.com/amidmm/MyChain/Bundle"
	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Initial"
	"github.com/amidmm/MyChain/Messages"
	"github.com/amidmm/MyChain/Packet"
	"github.com/amidmm/MyChain/Rep"
	"github.com/amidmm/MyChain/Tangle"
	"github.com/amidmm/MyChain/Transaction"
	"github.com/amidmm/MyChain/Weak"
)

func Validate(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) (bool, error) {
	if p.PacketType == msg.Packet_BLOCK {
		log.Println("\033[31m new block\033[0m")
		ok, err := ValidateBlock(p, bc, t)
		if err != nil || !ok {
			return false, err
		}
		for _, req := range p.GetBlockData().Reqs {
			Weak.AddNewWeakReq(req, p)
		}
		err = bc.AddBlock(p)
		if err != nil {
			return false, nil
		}
		log.Println("\033[31m block done\033[0m")
	} else {
		if p.PacketType == msg.Packet_WEAKREQ {
			if ok, err := Weak.ValidateWeakReq(p); err != nil || !ok {
				return false, err
			}
		} else if p.PacketType == msg.Packet_REP {
			if ok, err := Rep.ValidateRep(p, t); err != nil || !ok {
				return false, err
			}
		} else if p.PacketType == msg.Packet_INITIAL {
			if ok, err := Initial.ValidateInitial(p, t); err != nil || !ok {
				return false, err
			}
		}
		t.AddBundle(p, false)
	}

	return true, nil
}

func ValidateBlock(p *msg.Packet, bc *Blockchain.Blockchain, t *Tangle.Tangle) (bool, error) {
	switch p.Data.(type) {
	case *msg.Packet_BlockData:
		b := p.GetBlockData()
		if v, err := ValidateWeak(p); err != nil || !v {
			return false, err
		}
		if bun, err := ValidatePacketHashs(b, bc, t); err != nil || !bun {
			return false, err
		}
		san, err := ValidateSanity(b)
		// NotImplemented
		// if err != nil {
		// 	return false, err
		// }
		if !san {
			return false, nil
		}
		if coin, err := ValidateCoinbase(b, bc); err != nil || !coin {
			return false, nil
		}
		// Remove after implementation
		err = err
	default:
		return false, Consts.ErrNotABlock
	}
	return true, nil
}

func ValidateWeak(b *msg.Packet) (bool, error) {
	for _, v := range b.GetBlockData().Reqs {
		if t, err := Weak.ValidateWeakReq(v); err != nil || !t {
			return false, err
		}
	}
	return true, nil
}
func ValidatePacketHashs(b *msg.Block, bc *Blockchain.Blockchain, t *Tangle.Tangle) (bool, error) {
	tipCount := make(map[string]int)
	for _, h := range b.PacketHashs {
		p, _ := Packet.GetPacket(h.Hash)
		if bytes.Compare(p.CurrentBlockHash, bc.Tip.Hash) != 0 {
			return false, Consts.ErrWrongBlockHash
		}
		if v, err := Bundle.ValidateBundle(p.GetBundleData(), true); err != nil || !v {
			return false, err
		}
		if v, err := t.ValidateVerify(p, false); err != nil || !v {
			return false, err
		}
		tipCount[string(p.GetBundleData().Verify1)]++
		tipCount[string(p.GetBundleData().Verify2)]++
		if tipCount[string(p.GetBundleData().Verify1)] > 1 || tipCount[string(p.GetBundleData().Verify2)] > 1 {
			return false, nil
		}
	}
	return true, nil
}
func ValidateSanity(b *msg.Block) (bool, error) {
	return true, Consts.ErrNotImplemented
}
func ValidateCoinbase(b *msg.Block, bc *Blockchain.Blockchain) (bool, error) {

	if t, err := Transaction.ValidateTx(b.Coinbase, true); err != nil || !t {
		return false, err
	}
	coinBase, _ := bc.ExpectedBlockReward()
	if b.Coinbase.Value > coinBase {
		return false, Consts.ErrWrongBlockRewared
	}
	return true, nil
}

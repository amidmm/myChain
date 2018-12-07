package PoW

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/gogo/protobuf/proto"

	"golang.org/x/crypto/sha3"
)

type pow struct {
	rawData []byte
	target  uint32
}

func SetPoW(ctx context.Context, p *msg.Packet, target uint32) error {
	sign := p.Sign
	p.Nonce = []byte{}
	p.Sign = []byte{}
	p.Hash = []byte{}
	p.Diff = target
	data, err := proto.Marshal(p)
	if err != nil {
		return err
	}
	nonce, err := calculatePoW(ctx, data, target)
	if err != nil {
		return err
	}
	p.Nonce = nonce
	p.Sign = sign
	return nil
}

// calcuate the PoW of data
// this is done in Data + nonce formate
func calculatePoW(ctx context.Context, data []byte, target uint32) ([]byte, error) {
	// 512 as we are using sha3_512
	if target > 512 {
		return []byte{}, Consts.ErrWrongTarget
	}
	diff := big.NewInt(1)
	var nonce big.Int
	var hashInt big.Int
	one := big.NewInt(1)
	done := make(chan bool, 1)
	diff.Lsh(diff, uint(512-target))
	fmt.Println("Mining \"" + hex.EncodeToString(sha3.New224().Sum(data)) + "\" with target " + string(target))
	for {
		raw := bytes.Join([][]byte{
			data,
			nonce.Bytes(),
		}, []byte{})
		hash := sha3.New512()
		hash.Write(raw)
		hashInt.SetBytes(hash.Sum(nil))
		if hashInt.Cmp(diff) == -1 {
			break
		} else {
			nonce = *nonce.Add(&nonce, one)
			done <- false
		}
		select {
		case <-ctx.Done():
			return []byte{}, Consts.ErrCanceled
		case <-done:
			continue
		}
	}
	return nonce.Bytes(), nil
}

//Validate PoW of given Packet
func ValidatePoW(data msg.Packet, target uint32) (bool, error) {
	var hashInt big.Int
	diff := big.NewInt(1)
	// 512 as we are using sha3_512
	if target > 512 {
		return false, Consts.ErrWrongTarget
	}
	nonce := data.Nonce
	sign := data.Sign
	prevHash := data.Hash
	data.Nonce = []byte{}
	data.Sign = []byte{}
	data.Hash = []byte{}
	raw, err := proto.Marshal(&data)
	if err != nil {
		log.Println(err)
		return false, err
	}
	raw = bytes.Join([][]byte{
		raw,
		nonce,
	}, []byte{})
	hash := sha3.New512()
	hash.Write(raw)
	hashInt.SetBytes(hash.Sum(nil))
	diff.Lsh(diff, uint(512-target))
	if hashInt.Cmp(diff) == -1 {
		data.Sign = sign
		data.Nonce = nonce
		data.Hash = prevHash
		return true, nil
	} else {
		data.Sign = sign
		data.Nonce = nonce
		data.Hash = prevHash
		return false, nil
	}
}

func SetHash(p *msg.Packet) error {
	hash := sha3.New512()
	raw, err := proto.Marshal(p)
	if err != nil {
		return err
	}
	hash.Write(raw)
	p.Hash = hash.Sum(nil)
	return nil
}

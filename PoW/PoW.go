package PoW

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"math/big"
	"strconv"

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
	p.Nonce = nil
	p.Hash = nil
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
	return nil
}

// calculatePoW calcuates the PoW of data
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
	IdHash := sha1.New()
	IdHash.Write(data)
	hashForLog := hex.EncodeToString(IdHash.Sum(nil))
	log.Println("Mining \"" + hex.EncodeToString(IdHash.Sum(nil)) + "\" with target " + strconv.Itoa(int(target)))
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
	log.Println("Found nonce for " + hashForLog + " :" + nonce.String())
	return nonce.Bytes(), nil
}

// ValidatePoW validates PoW of given Packet
func ValidatePoW(data msg.Packet, target uint32) (bool, error) {
	var hashInt big.Int
	diff := big.NewInt(1)
	// 512 as we are using sha3_512
	if target > 512 {
		return false, Consts.ErrWrongTarget
	}
	nonce := data.Nonce
	prevHash := data.Hash
	data.Nonce = nil
	data.Hash = nil
	defer func() {
		data.Nonce = nonce
		data.Hash = prevHash
	}()
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
		return true, nil
	} else {
		return false, nil
	}
}

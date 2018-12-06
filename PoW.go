package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/amidmm/MyChain/Messages"
	"github.com/gogo/protobuf/proto"

	"golang.org/x/crypto/sha3"
)

type pow struct {
	rawData []byte
	target  uint32
}

// calcuate the PoW of data
// this is done in Data + nonce formate
func CalculatePoW(ctx context.Context, data []byte, target uint32) ([]byte, error) {
	// 512 as we are using sha3_512
	if target > 512 {
		return []byte{}, errors.New("wrong target value")
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
			return []byte{}, errors.New("calculation canceled")
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
		return false, errors.New("Wrong target value!")
	}
	nonce := data.Nonce
	sign := data.Sign
	data.Nonce = []byte{}
	data.Sign = []byte{}
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
		return true, nil
	} else {
		data.Sign = sign
		data.Nonce = nonce
		return false, nil
	}
}

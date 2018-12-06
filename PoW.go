package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

type pow struct {
	rawData []byte
	target  uint32
}

func CalculatePoW(ctx context.Context, data []byte, target uint32) []byte {
	if target > 512 {
		panic("Wrong target value!")
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
		fmt.Println(nonce.String())
		select {
		case <-ctx.Done():
			return nil
		case <-done:
			continue
		}
	}
	return nonce.Bytes()
}

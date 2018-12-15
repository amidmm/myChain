package Consts

import (
	"errors"
)

const Name = "myChain"
const BlockchainDB string = "Data/blockchain"
const TangleDB string = "Data/tangle"
const UTXODB string = "Data/UTXO"
const UserDB = "Data/user_%s"
const PoWLimit uint32 = 15

var (
	ErrBlockchainExists = errors.New(Name + " blockchain exists")
	ErrWrongTarget      = errors.New("wrong target value")
	ErrCanceled         = errors.New("calculation canceled")
	ErrWrongHeight      = errors.New("height limitation exceeded")
	ErrRetargetRetriv   = errors.New("unable to obtain previous retarget block")
	ErrNotImplemented   = errors.New("not implemented")
	ErrNotABlock        = errors.New("wrong Packet type as last block")
)

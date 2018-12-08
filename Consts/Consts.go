package Consts

import (
	"errors"
)

const Name = "myChain"
const BlockchainDB string = "Data/blockchain.db"
const TangleDB string = "Data/tangle.db"

var (
	ErrBlockchainExists = errors.New(Name + " blockchain exists")
	ErrWrongTarget      = errors.New("wrong target value")
	ErrCanceled         = errors.New("calculation canceled")
	ErrWrongHeight      = errors.New("height limitation exceeded")
	ErrRetargetRetriv   = errors.New("unable to obtain previous retarget block")
)

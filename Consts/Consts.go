package Consts

import (
	"errors"
)

const Name = "myChain"
const BlockchainDB string = "Data/blockchain"
const TangleDB string = "Data/tangle"
const TangleRelations string = "Data/tangle_relation"
const TangleUnApproved string = "Data/tangle_unapproved"
const UTXODB string = "Data/UTXO"
const UserDB = "Data/user_%s"
const PoWLimit uint32 = 20

var (
	ErrBlockchainExists     = errors.New(Name + " blockchain exists")
	ErrWrongTarget          = errors.New("wrong target value")
	ErrCanceled             = errors.New("calculation canceled")
	ErrWrongHeight          = errors.New("height limitation exceeded")
	ErrRetargetRetriv       = errors.New("unable to obtain previous retarget block")
	ErrNotImplemented       = errors.New("not implemented")
	ErrNotABlock            = errors.New("wrong Packet type as last block")
	ErrWrongBlockHash       = errors.New("wrong hash of current block for packetHashes")
	ErrWrongBlockRewared    = errors.New("bigger block reward than expected")
	ErrInconsistantTangleDB = errors.New("tangle databases are not setup properly")
	ErrOldBlock             = errors.New("block is older than tip block")
)

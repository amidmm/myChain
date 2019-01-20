package Consts

import (
	"errors"
	"time"
)

const ClientVersion = "myChain/0.0.1"
const Name = "myChain"
const BlockchainDB string = "Data/blockchain"
const TangleDB string = "Data/tangle"
const TangleRelations string = "Data/tangle_relation"
const TangleUnApproved string = "Data/tangle_unapproved"
const UTXODB string = "Data/UTXO"
const UserDB string = "Data/user_%s"
const UserTips string = "Data/users_tips"
const UnsyncPool string = "Data/unsync_pool"
const PoWLimit uint32 = 20

//TangleInitPoWLinit PoW for initial packets in tangle
const TangleInitPoWLimit uint32 = 12
const VbcPoWLimit uint32 = 20
const TimestampBound time.Duration = 18000 * time.Second

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
	ErrDataIntegrity        = errors.New("error in data integrity")
	ErrTangleExists         = errors.New(Name + " tangle exists")
	ErrWrongParam           = errors.New("wrong parameter is provided")
)

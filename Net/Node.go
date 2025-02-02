package Net

import (
	"bufio"
	"context"
	"log"
	"sync"

	"github.com/amidmm/MyChain/Account"

	"github.com/amidmm/MyChain/Config"

	"github.com/amidmm/MyChain/Sync"

	"github.com/amidmm/MyChain/Messages"

	"github.com/amidmm/MyChain/Blockchain"
	"github.com/amidmm/MyChain/Bootstrap"
	"github.com/amidmm/MyChain/Tangle"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
)

type Node struct {
	host.Host
	*NetProtocol
	*MsgProtocol
}

var Ctx = context.Background()
var PacketChan chan *msg.Packet
var AdvertiserChan chan *msg.Packet
var Bc *Blockchain.Blockchain
var T *Tangle.Tangle
var User *Account.User
var deadLockRand int64
var deadLockDone chan bool

func init() {
	var err error
	Ctx, PacketChan, Bc, T, User, AdvertiserChan, err = Bootstrap.Run()
	if err != nil {
		log.Panicln("Can't start Node")
	}
	go func() {
		Sync.User = User
		Sync.IncomingPacket(Ctx, PacketChan, Bc, T, AdvertiserChan)
	}()
	deadLockDone = make(chan bool)
}

func NewNode(host host.Host) *Node {
	node := &Node{Host: host}
	node.NetProtocol = NewNetProtocol(node)
	node.MsgProtocol = NewMsgProtocol(node)
	Config.This = node.Host.Addrs()
	go node.Advertiser()
	return node
}

var lock sync.Mutex

func (n *Node) SendPacket(data proto.Message, s inet.Stream) (bool, error) {
	lock.Lock()
	defer lock.Unlock()
	writer := bufio.NewWriter(s)
	enc := protobufCodec.Multicodec(nil).Encoder(writer)
	if err := enc.Encode(data); err != nil {
		return false, err
	}
	if err := writer.Flush(); err != nil {
		return false, err
	}
	return true, nil
}

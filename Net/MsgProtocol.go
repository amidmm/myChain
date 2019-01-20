package Net

import (
	"bufio"
	"log"
	"math/rand"
	"time"

	"github.com/amidmm/MyChain/Sync"

	"github.com/amidmm/MyChain/Messages"
	inet "github.com/libp2p/go-libp2p-net"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
)

const MsgProto = "/MsgProtocol/0.0.1"

type MsgProtocol struct {
	Node    *Node
	MsgChan chan *msg.Packet
}

func NewMsgProtocol(node *Node) *MsgProtocol {
	mp := &MsgProtocol{Node: node, MsgChan: make(chan *msg.Packet)}
	node.SetStreamHandler(MsgProto, mp.onMsg)
	return mp
}

func (mp *MsgProtocol) onMsg(s inet.Stream) {
	if Sync.SyncMode {
		rand.Seed(time.Now().Unix())
		deadLockRand = rand.Int63()
	}
	data := &msg.Packet{}
	decoder := protobufCodec.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		log.Println("\033[31m onMsg " + err.Error() + "\033[31m")
		return
	}
	//TODO: add a new channel to handel other peers packets while in sync mode
	PacketChan <- data
}

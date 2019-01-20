package Net

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/amidmm/MyChain/Sync"

	"github.com/amidmm/MyChain/Blockchain"

	"github.com/libp2p/go-libp2p-peer"

	"github.com/golang/protobuf/ptypes"

	"github.com/golang/protobuf/proto"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"

	"github.com/amidmm/MyChain/Config"
	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/NetMessages"
	crypto "github.com/libp2p/go-libp2p-crypto"
	inet "github.com/libp2p/go-libp2p-net"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

const Ping = "/Net/ping/0.0.1"
const Pong = "/Net/pong/0.0.1"
const FindNeighbourReq = "/Net/findNeighbour/0.0.1"
const Neighbour = "/Net/listNeighbour/0.0.1"
const SyncReq = "/Net/syncReq/0.0.1"
const SyncAck = "Net/sync/0.0.1"
const SyncStart = "Net/syncStart/0.0.1"
const SyncDone = "Net/syncDone/0.0.1"

var syncLock sync.Mutex
var bestSync peer.ID
var bestSyncHash []byte
var genericLock sync.Mutex

type NetProtocol struct {
	Node *Node
	Req  map[string]*NetMessages.NetPacket
}

func NewNetProtocol(node *Node) *NetProtocol {
	np := &NetProtocol{Node: node, Req: make(map[string]*NetMessages.NetPacket)}
	node.SetStreamHandler(Ping, np.onPing)
	node.SetStreamHandler(Pong, np.onPong)
	node.SetStreamHandler(FindNeighbourReq, np.onFindNeighbourReq)
	node.SetStreamHandler(Neighbour, np.onNeighbour)
	node.SetStreamHandler(SyncReq, np.onSyncReq)
	node.SetStreamHandler(SyncAck, np.onSyncAck)
	node.SetStreamHandler(SyncStart, np.onSyncStart)
	node.SetStreamHandler(SyncDone, np.onSyncDone)
	return np
}

func (np *NetProtocol) onPing(s inet.Stream) {
	log.Println("\033[33m onPing: received a new ping\033[0m")
	data, err := decodePacket(s)
	if err != nil {
		log.Println("\033[31m onPing: error parsing ping\033[0m")
		return
	}
	if ok := ValidateNetMsg(data); !ok {
		log.Println("\033[31m onPing: error parsing ping\033[0m")
		return
	}
	isInSet := false
	for _, v := range Config.This {
		if bytes.Equal(v.Bytes(), data.GetPingData().To) {
			isInSet = true
		}
	}
	if !isInSet {
		log.Println("\033[31m onPing: error parsing ping\033[0m")
	}
	data.Sign = nil
	raw, _ := proto.Marshal(data)
	hash := sha1.New()
	hash.Write(raw)
	pongPacket := EmptyNetMsg(NetMessages.NetPacket_PONG, nil, false, np)
	pd := &NetMessages.Pong{To: s.Conn().RemoteMultiaddr().Bytes(), ReplyTok: hash.Sum(nil)}
	pongPacket.Data = &NetMessages.NetPacket_PongData{pd}
	pongPacket.Sign = nil
	raw, _ = proto.Marshal(pongPacket)
	pongPacket.Sign, _ = np.Node.Peerstore().PrivKey(np.Node.ID()).Sign(raw)
	s, respErr := np.Node.NewStream(Ctx, s.Conn().RemotePeer(), Pong)
	if respErr != nil {
		log.Println("\033[31m onPing: error creating stream to" + s.Conn().RemotePeer().String() + "\t" + respErr.Error() + "\033[0m")
		return
	}
	if _, err := np.Node.SendPacket(pongPacket, s); err != nil {
		log.Println("\033[31m onPing: error sending pong " + respErr.Error() + "\033[0m")
		return
	}
	ipfsaddr, err := ma.NewMultiaddrBytes(data.NodeId)
	if err != nil {
		log.Fatalln(err)
	}
	info, err := peerstore.InfoFromP2pAddr(ipfsaddr)
	if err != nil {
		log.Fatalln(err)
	}
	np.Node.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
}

func (np *NetProtocol) onPong(s inet.Stream) {
	log.Println("\033[33m onPong: received a new pong\033[0m")
	data, err := decodePacket(s)
	if err != nil {
		log.Println("\033[31m onPong: error parsing pong\033[0m")
		return
	}
	if ok := ValidateNetMsg(data); !ok {
		log.Println("\033[31m onPong: error parsing pong\033[0m")
		return
	}
	// Check if it's for this host
	isInSet := false
	for _, v := range Config.This {
		if bytes.Equal(v.Bytes(), data.GetPongData().To) {
			isInSet = true
			break
		}
	}
	if !isInSet {
		log.Println("\033[31m onPong: error parsing pong\033[0m")
	}
	token := hex.EncodeToString(data.GetPongData().ReplyTok)
	if _, ok := np.Req[token]; !ok {
		log.Println("\033[31m onPong: not requested\033[0m")
		return
	}
	delete(np.Req, token)
}

func (np *NetProtocol) onFindNeighbourReq(s inet.Stream) {
	log.Println("\033[33m onFindNeighbourReq: received a new FindNeighbour\033[0m")
	data, err := decodePacket(s)
	if err != nil {
		log.Println("\033[31m onFindNeighbourReq: error parsing FindNeighbour\033[0m")
		return
	}
	if ok := ValidateNetMsg(data); !ok {
		log.Println("\033[31m onFindNeighbourReq: error parsing FindNeighbour\033[0m")
		return
	}
	neighbourPacket := EmptyNetMsg(NetMessages.NetPacket_NEIGHBOUR, nil, false, np)
	neighbourData := &NetMessages.Neighbours{}
	for _, v := range np.Node.Peerstore().Peers() {
		//TODO: remove loopbacks (in final phase)
		if v == np.Node.ID() {
			continue
		}
		if v == s.Conn().RemotePeer() {
			continue
		}
		x := np.Node.Peerstore().PeerInfo(v)
		y, err := peerstore.InfoToP2pAddrs(&x)
		if err != nil {
			continue
		}
		for _, w := range y {
			addr := &NetMessages.Node{Node: w.Bytes()}
			neighbourData.Nodes = append(neighbourData.Nodes, addr)
		}
	}
	neighbourPacket.Data = &NetMessages.NetPacket_NeighbourData{neighbourData}
	neighbourPacket.Sign = nil
	raw, _ := proto.Marshal(neighbourPacket)
	neighbourPacket.Sign, _ = np.Node.Peerstore().PrivKey(np.Node.ID()).Sign(raw)
	s, respErr := np.Node.NewStream(Ctx, s.Conn().RemotePeer(), Neighbour)
	if respErr != nil {
		log.Println("\033[31m onFindNeighbourReq: error creating stream to:" + s.Conn().RemotePeer().String() + "\t" + respErr.Error() + "\033[0m")
		return
	}
	if _, err := np.Node.SendPacket(neighbourPacket, s); err != nil {
		log.Println("\033[31m onFindNeighbourReq: error sending Neighbour " + respErr.Error() + "\033[0m")
		return
	}
}

func (np *NetProtocol) onNeighbour(s inet.Stream) {
	log.Println("\033[33m onNeighbour: received a new neighbour\033[0m")
	data, err := decodePacket(s)
	if err != nil {
		log.Println("\033[31m onNeighbour: error parsing neighbour\033[0m")
		return
	}
	if ok := ValidateNetMsg(data); !ok {
		log.Println("\033[31m onNeighbour: error parsing neighbour\033[0m")
		return
	}
	for _, addr := range data.GetNeighbourData().Nodes {
		if addr.Node != nil {
			remote, err := ma.NewMultiaddrBytes(addr.Node)
			if err != nil {
				continue
			}
			info, err := peerstore.InfoFromP2pAddr(remote)
			if info.ID == np.Node.ID() {
				continue
			}
			if info.ID == s.Conn().RemotePeer() {
				continue
			}
			np.Node.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
			if err != nil {
				continue
			}
			s, err := np.Node.NewStream(Ctx, info.ID, Ping)
			if err != nil {
				log.Println("\033[31m onNeighbour: error creating stream to: " + info.ID.String() + "\t" + err.Error() + "\033[0m")
				continue
			}
			if ok := np.SendPing(s); !ok {
				log.Println("\033[31m fail sending ping for advertised node\033[0m")
			}
		}
	}
}

func ValidateNetMsg(n *NetMessages.NetPacket) bool {
	var packetType NetMessages.NetPacketType = -1
	switch n.Data.(type) {
	case *NetMessages.NetPacket_PingData:
		packetType = NetMessages.NetPacket_PING
	case *NetMessages.NetPacket_PongData:
		packetType = NetMessages.NetPacket_PONG
	case *NetMessages.NetPacket_FindNeighbourReqData:
		packetType = NetMessages.NetPacket_FINDNEIGHBOURREQ
	case *NetMessages.NetPacket_NeighbourData:
		packetType = NetMessages.NetPacket_NEIGHBOUR
	case *NetMessages.NetPacket_SyncReqData:
		packetType = NetMessages.NetPacket_SYNCREQ
	case *NetMessages.NetPacket_SyncAckData:
		packetType = NetMessages.NetPacket_SYNCACK
	case *NetMessages.NetPacket_SyncStartData:
		packetType = NetMessages.NetPacket_SYNCSTART
	case *NetMessages.NetPacket_SyncDoneData:
		packetType = NetMessages.NetPacket_SYNCDONE
	}
	if packetType != n.PacketType {
		return false
	}
	lim := time.Now().Unix() - int64(Consts.TimestampBound.Seconds())
	if n.Timestamp.Seconds < lim {
		return false
	}
	if !bytes.Equal(n.Version, []byte(Consts.ClientVersion)) {
		return false
	}
	sign := n.Sign
	n.Sign = nil
	raw, _ := proto.Marshal(n)
	n.Sign = sign
	pk, err := crypto.UnmarshalPublicKey(n.NodePubKey)
	if err != nil {
		return false
	}
	if ok, err := pk.Verify(raw, sign); err != nil || !ok {
		return false
	}
	return true
}

func (np *NetProtocol) SendPing(s inet.Stream) bool {
	log.Println("\033[33m SendPing: sending a new ping\033[0m")
	pingPacket := EmptyNetMsg(NetMessages.NetPacket_PING, nil, false, np)
	pd := &NetMessages.Ping{To: s.Conn().RemoteMultiaddr().Bytes()}
	pingPacket.Data = &NetMessages.NetPacket_PingData{pd}
	pingPacket.Sign = nil
	raw, _ := proto.Marshal(pingPacket)
	pingPacket.Sign, _ = np.Node.Peerstore().PrivKey(np.Node.ID()).Sign(raw)
	hash := sha1.New()
	hash.Write(raw)
	np.Req[hex.EncodeToString(hash.Sum(nil))] = pingPacket
	if _, err := np.Node.SendPacket(pingPacket, s); err != nil {
		log.Println("\033[31m Ping: failed to send ping to:" + s.Conn().RemoteMultiaddr().String() + "\033[0m")
		return false
	}
	return true
}

func EmptyNetMsg(packType NetMessages.NetPacketType, id []byte, gossip bool, np *NetProtocol) *NetMessages.NetPacket {
	n := &NetMessages.NetPacket{}
	n.PacketType = packType
	n.Timestamp = ptypes.TimestampNow()
	n.Version = []byte(Consts.ClientVersion)
	n.Id = id
	n.Gossip = gossip
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", np.Node.ID().Pretty()))
	n.NodeId = np.Node.Addrs()[0].Encapsulate(hostAddr).Bytes()
	n.NodePubKey, _ = np.Node.Peerstore().PubKey(np.Node.ID()).Bytes()
	return n
}

func decodePacket(s inet.Stream) (*NetMessages.NetPacket, error) {
	data := &NetMessages.NetPacket{}
	decoder := protobufCodec.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (np *NetProtocol) SendFindNeighbour(s inet.Stream) bool {
	log.Println("\033[33m SendFindNeighbour: sending a new FindNeighbour\033[0m")
	FindNeighbourPacket := EmptyNetMsg(NetMessages.NetPacket_FINDNEIGHBOURREQ, nil, false, np)
	pd := &NetMessages.FindNeighbourReq{}
	FindNeighbourPacket.Data = &NetMessages.NetPacket_FindNeighbourReqData{pd}
	FindNeighbourPacket.Sign = nil
	raw, _ := proto.Marshal(FindNeighbourPacket)
	FindNeighbourPacket.Sign, _ = np.Node.Peerstore().PrivKey(np.Node.ID()).Sign(raw)
	if _, err := np.Node.SendPacket(FindNeighbourPacket, s); err != nil {
		log.Println("\033[31m SendFindNeighbour: failed to send FindNeighbour to:" + s.Conn().RemoteMultiaddr().String() + "\033[0m")
		return false
	}
	return true
}

func (np *NetProtocol) onSyncReq(s inet.Stream) {
	log.Println("\033[33m onSyncReq: received a new SyncReq\033[0m")
	data, err := decodePacket(s)
	if err != nil {
		log.Println("\033[31m onSyncReq: error parsing SyncReq\033[0m")
		return
	}
	if ok := ValidateNetMsg(data); !ok {
		log.Println("\033[31m onSyncReq: error parsing SyncReq\033[0m")
		return
	}
	if !bytes.Equal(data.GetSyncReqData().BcGenesisHash, Bc.GenesisHash) {
		log.Println("\033[31m onSyncReq: unidentical Blockchain genesis\033[0m")
		return
	}
	if !bytes.Equal(data.GetSyncReqData().TGenesisHash, T.GenesisHash1) || !bytes.Equal(data.GetSyncReqData().TGenesis2Hash, T.GenesisHash2) {
		log.Println("\033[31m onSyncReq: unidentical Tangle genesis\033[0m")
		return
	}
	syncAck := EmptyNetMsg(NetMessages.NetPacket_SYNCACK, nil, false, np)
	syncAckData := &NetMessages.SyncAck{CurrentBlockHash: Bc.Tip.Hash, CurrentBlockHeight: Bc.Tip.CurrentBlockNumber}
	syncAck.Data = &NetMessages.NetPacket_SyncAckData{syncAckData}
	syncAck.Sign = nil
	raw, _ := proto.Marshal(syncAck)
	syncAck.Sign, _ = np.Node.Peerstore().PrivKey(np.Node.ID()).Sign(raw)
	s, respErr := np.Node.NewStream(Ctx, s.Conn().RemotePeer(), SyncAck)
	if respErr != nil {
		log.Println("\033[31m onSyncReq: error creating stream to" + s.Conn().RemotePeer().String() + "\t" + respErr.Error() + "\033[0m")
		return
	}
	if _, err := np.Node.SendPacket(syncAck, s); err != nil {
		log.Println("\033[31m onSyncReq: error sending SyncAck " + respErr.Error() + "\033[0m")
		return
	}
}

func (np *NetProtocol) onSyncAck(s inet.Stream) {
	log.Println("\033[33m onSync: received a new SyncAck\033[0m")
	data, err := decodePacket(s)
	if err != nil {
		log.Println("\033[31m onSync: error parsing SyncAck\033[0m")
		return
	}
	if ok := ValidateNetMsg(data); !ok {
		log.Println("\033[31m onSync: error parsing SyncAck\033[0m")
		return
	}
	//TODO: handle forks much better
	genericLock.Lock()
	defer genericLock.Unlock()
	if data.GetSyncAckData().CurrentBlockHeight > Bc.Tip.CurrentBlockNumber {
		bestSync = s.Conn().RemotePeer()
		bestSyncHash = data.GetSyncAckData().CurrentBlockHash
	}
}

func (np *NetProtocol) SendSyncReq(s inet.Stream) bool {
	log.Println("\033[33m SendSyncReq: sending a new SyncReq\033[0m")
	SyncReqPacket := EmptyNetMsg(NetMessages.NetPacket_SYNCREQ, nil, false, np)
	pd := &NetMessages.SyncReq{BcGenesisHash: Bc.GenesisHash, TGenesisHash: T.GenesisHash1, TGenesis2Hash: T.GenesisHash2}
	SyncReqPacket.Data = &NetMessages.NetPacket_SyncReqData{pd}
	SyncReqPacket.Sign = nil
	raw, _ := proto.Marshal(SyncReqPacket)
	SyncReqPacket.Sign, _ = np.Node.Peerstore().PrivKey(np.Node.ID()).Sign(raw)
	if _, err := np.Node.SendPacket(SyncReqPacket, s); err != nil {
		log.Println("\033[31m SendSyncReq: failed to send SyncReq to:" + s.Conn().RemoteMultiaddr().String() + "\033[0m")
		return false
	}
	return true
}

func (np *NetProtocol) FindBestSync() (peer.ID, []byte, error) {
	syncLock.Lock()
	defer syncLock.Unlock()
	var wg sync.WaitGroup
	defer func() {
		bestSyncHash = []byte{}
	}()
	for _, peerAddr := range np.Node.Peerstore().PeersWithAddrs() {
		pa := peerAddr
		wg.Add(1)
		go func(addr peer.ID) {
			defer wg.Done()
			s, err := np.Node.NewStream(Ctx, pa, SyncReq)
			if err != nil {
				return
			}
			np.SendSyncReq(s)
		}(pa)
	}
	wg.Wait()
	if len(bestSyncHash) == 0 {
		return bestSync, nil, errors.New("no best peer")
	}
	tmpbestSync := bestSync
	tmpbestHash := bestSyncHash
	return tmpbestSync, tmpbestHash, nil
}

func (np *NetProtocol) onSyncStart(s inet.Stream) {
	log.Println("\033[33m onSyncStart: received a new SyncStart\033[0m")
	data, err := decodePacket(s)
	if err != nil {
		log.Println("\033[31m onSyncStart: error parsing SyncStart\033[0m")
		return
	}
	if ok := ValidateNetMsg(data); !ok {
		log.Println("\033[31m onSyncStart: error parsing SyncStart\033[0m")
		return
	}
	//TODO: handle tangle messages !important
	//TODO: change on fork handling
	p, err := Bc.Ancestor(data.GetSyncStartData().CurrentBlockHeight)
	if err != nil {
		log.Println("\033[31m onSyncStart: can't find ancestor\033[0m")
		return
	}
	if !bytes.Equal(p.Hash, data.GetSyncStartData().CurrentBlockHash) {
		log.Println("\033[31m onSyncStart: fork detected\033[0m")
		return
	}
	iter := Blockchain.BlockchainIterator{}
	iter.InitIter(Bc)
	pid := s.Conn().RemotePeer()
	s, err = np.Node.NewStream(Ctx, pid, MsgProto)
	if err != nil {
		log.Println("\033[31m onSyncStart: error creating stream to" + s.Conn().RemotePeer().String() + "\t" + err.Error() + "\033[0m")
		return
	}
	value, _ := iter.Value()
	if _, err := np.Node.SendPacket(value, s); err != nil {
		log.Println("\033[31m onSyncReq: error sending SyncAck " + err.Error() + "\033[0m")
		return
	}
	for iter.Prev() && data.GetSyncStartData().CurrentBlockHeight < value.CurrentBlockNumber {
		genericLock.Lock()
		value, _ = iter.Value()
		s, err = np.Node.NewStream(Ctx, pid, MsgProto)
		if err != nil {
			log.Println("\033[31m onSyncStart: error creating stream to" + s.Conn().RemotePeer().String() + "\t" + err.Error() + "\033[0m")
			return
		}
		if _, err := np.Node.SendPacket(value, s); err != nil {
			log.Println("\033[31m onSyncReq: error sending MsgPacket " + err.Error() + "\033[0m")
			return
		}
		genericLock.Unlock()
	}
	time.Sleep(5 * time.Second)
	donePacket := EmptyNetMsg(NetMessages.NetPacket_SYNCDONE, nil, false, np)
	dd := &NetMessages.SyncDone{}
	donePacket.Data = &NetMessages.NetPacket_SyncDoneData{dd}
	donePacket.Sign = nil
	raw, _ := proto.Marshal(donePacket)
	donePacket.Sign, _ = np.Node.Peerstore().PrivKey(np.Node.ID()).Sign(raw)
	s, err = np.Node.NewStream(Ctx, pid, SyncDone)
	if err != nil {
		log.Println("\033[31m onSyncStart: error creating stream to" + s.Conn().RemotePeer().String() + "\t" + err.Error() + "\033[0m")
		return
	}
	if _, err := np.Node.SendPacket(donePacket, s); err != nil {
		log.Println("\033[31m onSyncReq: error sending SyncDone " + err.Error() + "\033[0m")
		return
	}
}

func (np *NetProtocol) onSyncDone(s inet.Stream) {
	log.Println("\033[33m onSyncDone: received a new SyncDone\033[0m")
	data, err := decodePacket(s)
	if err != nil {
		log.Println("\033[31m onSyncDone: error parsing SyncDone\033[0m")
		return
	}
	if ok := ValidateNetMsg(data); !ok {
		log.Println("\033[31m onSyncDone: error parsing SyncDone\033[0m")
		return
	}
	Sync.SyncMode = false
	deadLockDone <- true
	syncLock.Unlock()
}

func (np *NetProtocol) SendSyncStart(s inet.Stream) bool {
	log.Println("\033[33m SendSyncStart: sending a new SyncStart\033[0m")
	SyncStartPacket := EmptyNetMsg(NetMessages.NetPacket_SYNCSTART, nil, false, np)
	pd := &NetMessages.SyncStart{CurrentBlockHash: Bc.Tip.Hash, CurrentBlockHeight: Bc.Tip.CurrentBlockNumber}
	SyncStartPacket.Data = &NetMessages.NetPacket_SyncStartData{pd}
	SyncStartPacket.Sign = nil
	raw, _ := proto.Marshal(SyncStartPacket)
	SyncStartPacket.Sign, _ = np.Node.Peerstore().PrivKey(np.Node.ID()).Sign(raw)
	if _, err := np.Node.SendPacket(SyncStartPacket, s); err != nil {
		log.Println("\033[31m SendSyncStart: failed to send SyncStart to:" + s.Conn().RemoteMultiaddr().String() + "\033[0m")
		return false
	}
	syncLock.Lock()
	Sync.SyncMode = true
	go deadLockDetect()
	return true
}

func deadLockDetect() {
	t := time.NewTimer(time.Second * 60)
	start := deadLockRand
	for {
		select {
		case <-t.C:
			if deadLockRand == start {
				Sync.SyncMode = false
				syncLock.Unlock()
				return
			}
		case <-deadLockDone:
			return
		}
	}
}

package Rep

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"

	"github.com/amidmm/MyChain/Transaction"

	"github.com/go-ethereum/crypto/sha3"
	"github.com/golang/protobuf/proto"

	"github.com/amidmm/MyChain/Tangle"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

//change statistics
var RepPool *leveldb.DB = nil

func init() {
	if RepPool == nil {
		s, err := storage.OpenFile(Consts.RepPool, false)
		if err != nil {
			log.Fatalln("unable to open WeakReqDB")
		}
		db, err := leveldb.Open(s, nil)
		if err != nil {
			log.Fatalln("unable to open WeakReqDB")
		}
		RepPool = db
	}
}

func ValidateRep(p *msg.Packet, t *Tangle.Tangle) (bool, error) {
	if p.GetRepData() == nil {
		return false, Consts.ErrWrongParam
	}
	if !bytes.Equal(p.GetRepData().Hash, GetHash(p.GetRepData())) {
		return false, errors.New("wrong hash for rep")
	}
	switch p.GetRepData().RepType {
	case msg.Rep_POPR:
		if p.GetRepData().GetPOPRData() == nil {
			return false, Consts.ErrWrongParam
		}
		if p.GetRepData().GetPOPRData().Certified != nil {
			var certRaw bytes.Buffer
			enc := gob.NewEncoder(&certRaw)
			enc.Encode(p.GetRepData().Nonce)
			enc.Encode(p.Addr)
			enc.Encode(p.Timestamp)
			pub, err := crypto.UnmarshalPublicKey(p.GetRepData().GetPOPRData().Certified.CertAddr)
			if err != nil {
				return false, err
			}
			result, _ := pub.Verify(certRaw.Bytes(), p.GetRepData().GetPOPRData().Certified.CertSign)
			if result {
				return true, nil
			}
			if p.GetRepData().Nonce == nil {
				return false, errors.New("wrong nonce")
			}
			return false, nil
		}
		return true, nil

	case msg.Rep_POPRA:
		if p.GetRepData().GetPOPRAData() == nil {
			return false, Consts.ErrWrongParam
		}
		poprRaw, err := t.DB.Get(bytes.Join(
			[][]byte{[]byte("b"), p.GetRepData().Ref}, []byte{}),
			nil)
		if err != nil {
			return false, nil
		}
		popr := &msg.Packet{}
		proto.Unmarshal(poprRaw, popr)
		if popr.PacketType != msg.Packet_REP || popr.GetRepData().RepType != msg.Rep_POPR {
			return false, errors.New("wrong ref")
		} else if !bytes.Equal(popr.GetRepData().Addr, p.Addr) || !bytes.Equal(popr.Addr, p.GetRepData().Addr) {
			return false, errors.New("wrong ref")
		}
		if !bytes.Equal(p.GetRepData().Nonce, popr.GetRepData().Nonce) {
			return false, errors.New("wrong nonce")
		}
		return true, nil

	case msg.Rep_AGREE:
		if p.GetRepData().GetAgreeData() == nil {
			return false, Consts.ErrWrongParam
		}
		popaRaw, err := t.DB.Get(bytes.Join(
			[][]byte{[]byte("b"), p.GetRepData().Ref}, []byte{}),
			nil)
		if err != nil {
			return false, nil
		}
		popa := &msg.Packet{}
		proto.Unmarshal(popaRaw, popa)
		if popa.PacketType != msg.Packet_REP || popa.GetRepData().RepType != msg.Rep_POPRA {
			return false, errors.New("wrong ref")
		} else if !bytes.Equal(popa.GetRepData().Addr, p.Addr) || !bytes.Equal(popa.Addr, p.GetRepData().Addr) {
			return false, errors.New("wrong ref")
		}
		if !bytes.Equal(p.GetRepData().Nonce, popa.GetRepData().Nonce) {
			return false, errors.New("wrong nonce")
		}
		med := false
		for _, m := range popa.GetRepData().GetPOPRAData().Mediator {
			if bytes.Equal(p.GetRepData().GetAgreeData().Mediator, m) {
				med = true
			}
		}
		if popa.GetRepData().GetPOPRAData().Mediator == nil && p.GetRepData().GetAgreeData().Mediator == nil {
			med = true
		}
		if !med {
			return false, errors.New("med not in popra list")
		}
		if popa.GetRepData().GetPOPRAData().LockOnly {
			if p.GetRepData().GetAgreeData().LockMoney == nil {
				return false, Consts.ErrWrongParam
			}
			if ok, err := Transaction.HandleLockBundle(p); err != nil || !ok {
				return false, errors.New("wrong lock bundle")
			}
		}
		return true, nil
	case msg.Rep_CANCEL:
		if p.GetRepData().GetCancelData() == nil {
			return false, Consts.ErrWrongParam
		}
	case msg.Rep_COMPLAINT:
		if p.GetRepData().GetComplaintData() == nil {
			return false, Consts.ErrWrongParam
		}
	case msg.Rep_MEDIATOR:
		if p.GetRepData().GetMediatorData() == nil {
			return false, Consts.ErrWrongParam
		}
	case msg.Rep_REVIEW:
		if p.GetRepData().GetReviewData() == nil {
			return false, Consts.ErrWrongParam
		}

	case msg.Rep_REVENGE:
		if p.GetRepData().GetRevengeData() == nil {
			return false, Consts.ErrWrongParam
		}
	default:
		return false, Consts.ErrWrongParam
	}
	return false, nil
}

func GetHash(rep *msg.Rep) []byte {
	rep.Hash = nil
	raw, err := proto.Marshal(rep)
	if err == nil {
		return nil
	}
	hash := sha3.New512()
	hash.Write(raw)
	return hash.Sum(nil)
}

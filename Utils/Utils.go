package Utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/amidmm/MyChain/Consts"
	"github.com/amidmm/MyChain/Messages"
	"github.com/gogo/protobuf/proto"
)

func MarshalPacketList(packets []*msg.Packet) ([]byte, error) {
	var buf bytes.Buffer
	for _, packet := range packets {
		num := make([]byte, 8)
		raw, err := proto.Marshal(packet)
		if err != nil {
			return nil, err
		}
		binary.PutUvarint(num, uint64(len(raw)))
		buf.Write(num)
		buf.Write(raw)
	}
	return buf.Bytes(), nil
}

func UnMarshalPacketList(data []byte) ([]*msg.Packet, error) {
	var msgs []*msg.Packet
	for len(data) != 0 {
		num, _ := binary.Uvarint(data)
		data = data[8:]
		raw := data[:num]
		m := &msg.Packet{}
		err := proto.Unmarshal(raw, m)
		if err != nil {
			return nil, err
		}
		data = data[num:]
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func AppendMarshal(data []byte, p *msg.Packet) ([]byte, error) {
	var packets []*msg.Packet
	var buf bytes.Buffer
	buf.Write(data)
	packets = append(packets, p)
	newRaw, err := MarshalPacketList(packets)
	if err != nil {
		return nil, err
	}
	buf.Write(newRaw)
	return buf.Bytes(), nil
}

func MarshalSha3List(list [][]byte) []byte {
	var buf bytes.Buffer
	for _, l := range list {
		buf.Write(l)
	}
	return buf.Bytes()
}

func AppendMarshalSha3(data, hash []byte) ([]byte, error) {
	var buf bytes.Buffer
	// 512 / 8 = 64 as we use sha3_512
	if len(data)%64 != 0 {
		return nil, Consts.ErrDataIntegrity
	}
	buf.Write(data)
	buf.Write(hash)
	return buf.Bytes(), nil
}

func UnMarshalSha3List(data []byte) ([][]byte, error) {
	// 512 / 8 = 64 as we use sha3_512
	if len(data)%64 != 0 {
		return nil, Consts.ErrDataIntegrity
	} else if len(data) == 0 {
		return nil, errors.New("empty list")
	}
	var SHA3list [][]byte
	for len(data) != 0 {
		SHA3list = append(SHA3list, data[:64])
		data = data[64:]
	}
	return SHA3list, nil
}

func HashTo512(bigInt []byte) []byte {
	if len(bigInt) >= 64 {
		return bigInt
	}
	padding := 64 - len(bigInt)
	pad := []byte(strings.Repeat("\x00", padding))
	return bytes.Join(
		[][]byte{pad, bigInt}, []byte{})
}

func RemoveExistingDB() {
	os.RemoveAll(Consts.BlockchainDB)
	os.RemoveAll(Consts.TangleDB)
	os.RemoveAll(Consts.TangleRelations)
	os.RemoveAll(Consts.TangleUnApproved)
	os.RemoveAll(Consts.UTXODB)
	os.RemoveAll(Consts.UserTips)
}

// MarshalPacketCounter marshals the counters for virtual blockchains inn tangle
func MarshalPacketCounter(data map[uint64]uint64, lastBlockToVerify uint64) ([]byte, error) {
	CheckLengthPacketCounter(data, lastBlockToVerify)
	var buf bytes.Buffer
	for i, v := range data {
		keyByte := make([]byte, 8)
		valueByte := make([]byte, 8)
		binary.PutUvarint(keyByte, i)
		binary.PutUvarint(valueByte, v)
		_, err := buf.Write(keyByte)
		if err != nil {
			return nil, err
		}
		_, err = buf.Write(valueByte)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func CheckLengthPacketCounter(data map[uint64]uint64, lastBlockToVerify uint64) {
	if len(data) <= int(lastBlockToVerify) {
		return
	}
	keys := reflect.ValueOf(data).MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Uint() < keys[j].Uint()
	})
	for i := 0; i <= len(data)-int(lastBlockToVerify); i++ {
		delete(data, keys[i].Uint())
	}
}

func UnMarshalPacketCounter(raw []byte) map[uint64]uint64 {
	counter := 8
	data := make(map[uint64]uint64)
	for len(raw) != 0 {
		key, _ := binary.Uvarint(raw)
		value, _ := binary.Uvarint(raw[counter:])
		data[key] = value
		raw = raw[2*counter:]
	}
	return data
}

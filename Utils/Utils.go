package Utils

import (
	"bytes"
	"encoding/binary"
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

package Utils

import (
	"bytes"
	"encoding/binary"

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

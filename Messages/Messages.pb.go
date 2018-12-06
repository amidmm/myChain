// Code generated by protoc-gen-go. DO NOT EDIT.
// source: Messages/Messages.proto

package msg

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type PacketType int32

const (
	Packet_BUNDLE      PacketType = 0
	Packet_BLOCK       PacketType = 1
	Packet_REP         PacketType = 2
	Packet_WEAKREQ     PacketType = 3
	Packet_SANITYCHECK PacketType = 4
)

var PacketType_name = map[int32]string{
	0: "BUNDLE",
	1: "BLOCK",
	2: "REP",
	3: "WEAKREQ",
	4: "SANITYCHECK",
}

var PacketType_value = map[string]int32{
	"BUNDLE":      0,
	"BLOCK":       1,
	"REP":         2,
	"WEAKREQ":     3,
	"SANITYCHECK": 4,
}

func (x PacketType) String() string {
	return proto.EnumName(PacketType_name, int32(x))
}

func (PacketType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{0, 0}
}

type BundleType int32

const (
	Bundle_POWERFUL BundleType = 0
	Bundle_WEAK     BundleType = 1
)

var BundleType_name = map[int32]string{
	0: "POWERFUL",
	1: "WEAK",
}

var BundleType_value = map[string]int32{
	"POWERFUL": 0,
	"WEAK":     1,
}

func (x BundleType) String() string {
	return proto.EnumName(BundleType_name, int32(x))
}

func (BundleType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{2, 0}
}

type RepType int32

const (
	Rep_POPR    RepType = 0
	Rep_POPRA   RepType = 1
	Rep_REVIEW  RepType = 2
	Rep_INITIAL RepType = 3
)

var RepType_name = map[int32]string{
	0: "POPR",
	1: "POPRA",
	2: "REVIEW",
	3: "INITIAL",
}

var RepType_value = map[string]int32{
	"POPR":    0,
	"POPRA":   1,
	"REVIEW":  2,
	"INITIAL": 3,
}

func (x RepType) String() string {
	return proto.EnumName(RepType_name, int32(x))
}

func (RepType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{5, 0}
}

type Packet struct {
	PacketType         PacketType           `protobuf:"varint,1,opt,name=packet_type,json=packetType,proto3,enum=msg.PacketType" json:"packet_type,omitempty"`
	Timestamp          *timestamp.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Prev               []byte               `protobuf:"bytes,3,opt,name=prev,proto3" json:"prev,omitempty"`
	Nonce              []byte               `protobuf:"bytes,4,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Diff               uint32               `protobuf:"varint,5,opt,name=diff,proto3" json:"diff,omitempty"`
	Hash               []byte               `protobuf:"bytes,6,opt,name=hash,proto3" json:"hash,omitempty"`
	CurrentBlockNumber uint64               `protobuf:"varint,7,opt,name=current_block_number,json=currentBlockNumber,proto3" json:"current_block_number,omitempty"`
	Addr               []byte               `protobuf:"bytes,13,opt,name=addr,proto3" json:"addr,omitempty"`
	Sign               []byte               `protobuf:"bytes,14,opt,name=sign,proto3" json:"sign,omitempty"`
	// Types that are valid to be assigned to Data:
	//	*Packet_BlockData
	//	*Packet_BundleData
	//	*Packet_RepData
	//	*Packet_WeakData
	//	*Packet_SanityData
	Data                 isPacket_Data `protobuf_oneof:"data"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Packet) Reset()         { *m = Packet{} }
func (m *Packet) String() string { return proto.CompactTextString(m) }
func (*Packet) ProtoMessage()    {}
func (*Packet) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{0}
}

func (m *Packet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Packet.Unmarshal(m, b)
}
func (m *Packet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Packet.Marshal(b, m, deterministic)
}
func (m *Packet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Packet.Merge(m, src)
}
func (m *Packet) XXX_Size() int {
	return xxx_messageInfo_Packet.Size(m)
}
func (m *Packet) XXX_DiscardUnknown() {
	xxx_messageInfo_Packet.DiscardUnknown(m)
}

var xxx_messageInfo_Packet proto.InternalMessageInfo

func (m *Packet) GetPacketType() PacketType {
	if m != nil {
		return m.PacketType
	}
	return Packet_BUNDLE
}

func (m *Packet) GetTimestamp() *timestamp.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *Packet) GetPrev() []byte {
	if m != nil {
		return m.Prev
	}
	return nil
}

func (m *Packet) GetNonce() []byte {
	if m != nil {
		return m.Nonce
	}
	return nil
}

func (m *Packet) GetDiff() uint32 {
	if m != nil {
		return m.Diff
	}
	return 0
}

func (m *Packet) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Packet) GetCurrentBlockNumber() uint64 {
	if m != nil {
		return m.CurrentBlockNumber
	}
	return 0
}

func (m *Packet) GetAddr() []byte {
	if m != nil {
		return m.Addr
	}
	return nil
}

func (m *Packet) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

type isPacket_Data interface {
	isPacket_Data()
}

type Packet_BlockData struct {
	BlockData *Block `protobuf:"bytes,8,opt,name=blockData,proto3,oneof"`
}

type Packet_BundleData struct {
	BundleData *Bundle `protobuf:"bytes,9,opt,name=bundleData,proto3,oneof"`
}

type Packet_RepData struct {
	RepData *Rep `protobuf:"bytes,10,opt,name=repData,proto3,oneof"`
}

type Packet_WeakData struct {
	WeakData *WeakReq `protobuf:"bytes,11,opt,name=weakData,proto3,oneof"`
}

type Packet_SanityData struct {
	SanityData *SanityCheck `protobuf:"bytes,12,opt,name=sanityData,proto3,oneof"`
}

func (*Packet_BlockData) isPacket_Data() {}

func (*Packet_BundleData) isPacket_Data() {}

func (*Packet_RepData) isPacket_Data() {}

func (*Packet_WeakData) isPacket_Data() {}

func (*Packet_SanityData) isPacket_Data() {}

func (m *Packet) GetData() isPacket_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Packet) GetBlockData() *Block {
	if x, ok := m.GetData().(*Packet_BlockData); ok {
		return x.BlockData
	}
	return nil
}

func (m *Packet) GetBundleData() *Bundle {
	if x, ok := m.GetData().(*Packet_BundleData); ok {
		return x.BundleData
	}
	return nil
}

func (m *Packet) GetRepData() *Rep {
	if x, ok := m.GetData().(*Packet_RepData); ok {
		return x.RepData
	}
	return nil
}

func (m *Packet) GetWeakData() *WeakReq {
	if x, ok := m.GetData().(*Packet_WeakData); ok {
		return x.WeakData
	}
	return nil
}

func (m *Packet) GetSanityData() *SanityCheck {
	if x, ok := m.GetData().(*Packet_SanityData); ok {
		return x.SanityData
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Packet) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Packet_OneofMarshaler, _Packet_OneofUnmarshaler, _Packet_OneofSizer, []interface{}{
		(*Packet_BlockData)(nil),
		(*Packet_BundleData)(nil),
		(*Packet_RepData)(nil),
		(*Packet_WeakData)(nil),
		(*Packet_SanityData)(nil),
	}
}

func _Packet_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Packet)
	// data
	switch x := m.Data.(type) {
	case *Packet_BlockData:
		b.EncodeVarint(8<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.BlockData); err != nil {
			return err
		}
	case *Packet_BundleData:
		b.EncodeVarint(9<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.BundleData); err != nil {
			return err
		}
	case *Packet_RepData:
		b.EncodeVarint(10<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.RepData); err != nil {
			return err
		}
	case *Packet_WeakData:
		b.EncodeVarint(11<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.WeakData); err != nil {
			return err
		}
	case *Packet_SanityData:
		b.EncodeVarint(12<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.SanityData); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Packet.Data has unexpected type %T", x)
	}
	return nil
}

func _Packet_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Packet)
	switch tag {
	case 8: // data.blockData
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Block)
		err := b.DecodeMessage(msg)
		m.Data = &Packet_BlockData{msg}
		return true, err
	case 9: // data.bundleData
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Bundle)
		err := b.DecodeMessage(msg)
		m.Data = &Packet_BundleData{msg}
		return true, err
	case 10: // data.repData
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Rep)
		err := b.DecodeMessage(msg)
		m.Data = &Packet_RepData{msg}
		return true, err
	case 11: // data.weakData
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(WeakReq)
		err := b.DecodeMessage(msg)
		m.Data = &Packet_WeakData{msg}
		return true, err
	case 12: // data.sanityData
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(SanityCheck)
		err := b.DecodeMessage(msg)
		m.Data = &Packet_SanityData{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Packet_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Packet)
	// data
	switch x := m.Data.(type) {
	case *Packet_BlockData:
		s := proto.Size(x.BlockData)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Packet_BundleData:
		s := proto.Size(x.BundleData)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Packet_RepData:
		s := proto.Size(x.RepData)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Packet_WeakData:
		s := proto.Size(x.WeakData)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Packet_SanityData:
		s := proto.Size(x.SanityData)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type Block struct {
	Reqs                 []*WeakReq     `protobuf:"bytes,1,rep,name=reqs,proto3" json:"reqs,omitempty"`
	Bundles              []*Bundle      `protobuf:"bytes,2,rep,name=bundles,proto3" json:"bundles,omitempty"`
	Sanities             []*SanityCheck `protobuf:"bytes,3,rep,name=sanities,proto3" json:"sanities,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Block) Reset()         { *m = Block{} }
func (m *Block) String() string { return proto.CompactTextString(m) }
func (*Block) ProtoMessage()    {}
func (*Block) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{1}
}

func (m *Block) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Block.Unmarshal(m, b)
}
func (m *Block) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Block.Marshal(b, m, deterministic)
}
func (m *Block) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Block.Merge(m, src)
}
func (m *Block) XXX_Size() int {
	return xxx_messageInfo_Block.Size(m)
}
func (m *Block) XXX_DiscardUnknown() {
	xxx_messageInfo_Block.DiscardUnknown(m)
}

var xxx_messageInfo_Block proto.InternalMessageInfo

func (m *Block) GetReqs() []*WeakReq {
	if m != nil {
		return m.Reqs
	}
	return nil
}

func (m *Block) GetBundles() []*Bundle {
	if m != nil {
		return m.Bundles
	}
	return nil
}

func (m *Block) GetSanities() []*SanityCheck {
	if m != nil {
		return m.Sanities
	}
	return nil
}

type Bundle struct {
	BundleType           BundleType `protobuf:"varint,1,opt,name=bundle_type,json=bundleType,proto3,enum=msg.BundleType" json:"bundle_type,omitempty"`
	Hash                 []byte     `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	Transactions         []*Tx      `protobuf:"bytes,3,rep,name=transactions,proto3" json:"transactions,omitempty"`
	Verify1              []byte     `protobuf:"bytes,4,opt,name=verify1,proto3" json:"verify1,omitempty"`
	Verify2              []byte     `protobuf:"bytes,5,opt,name=verify2,proto3" json:"verify2,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Bundle) Reset()         { *m = Bundle{} }
func (m *Bundle) String() string { return proto.CompactTextString(m) }
func (*Bundle) ProtoMessage()    {}
func (*Bundle) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{2}
}

func (m *Bundle) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Bundle.Unmarshal(m, b)
}
func (m *Bundle) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Bundle.Marshal(b, m, deterministic)
}
func (m *Bundle) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Bundle.Merge(m, src)
}
func (m *Bundle) XXX_Size() int {
	return xxx_messageInfo_Bundle.Size(m)
}
func (m *Bundle) XXX_DiscardUnknown() {
	xxx_messageInfo_Bundle.DiscardUnknown(m)
}

var xxx_messageInfo_Bundle proto.InternalMessageInfo

func (m *Bundle) GetBundleType() BundleType {
	if m != nil {
		return m.BundleType
	}
	return Bundle_POWERFUL
}

func (m *Bundle) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Bundle) GetTransactions() []*Tx {
	if m != nil {
		return m.Transactions
	}
	return nil
}

func (m *Bundle) GetVerify1() []byte {
	if m != nil {
		return m.Verify1
	}
	return nil
}

func (m *Bundle) GetVerify2() []byte {
	if m != nil {
		return m.Verify2
	}
	return nil
}

type Tx struct {
	Hash                 []byte   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	BundleHash           []byte   `protobuf:"bytes,6,opt,name=bundle_hash,json=bundleHash,proto3" json:"bundle_hash,omitempty"`
	RefTx                []byte   `protobuf:"bytes,2,opt,name=refTx,proto3" json:"refTx,omitempty"`
	Sign                 []byte   `protobuf:"bytes,3,opt,name=sign,proto3" json:"sign,omitempty"`
	Value                int64    `protobuf:"varint,4,opt,name=value,proto3" json:"value,omitempty"`
	Tag                  []byte   `protobuf:"bytes,5,opt,name=tag,proto3" json:"tag,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Tx) Reset()         { *m = Tx{} }
func (m *Tx) String() string { return proto.CompactTextString(m) }
func (*Tx) ProtoMessage()    {}
func (*Tx) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{3}
}

func (m *Tx) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Tx.Unmarshal(m, b)
}
func (m *Tx) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Tx.Marshal(b, m, deterministic)
}
func (m *Tx) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Tx.Merge(m, src)
}
func (m *Tx) XXX_Size() int {
	return xxx_messageInfo_Tx.Size(m)
}
func (m *Tx) XXX_DiscardUnknown() {
	xxx_messageInfo_Tx.DiscardUnknown(m)
}

var xxx_messageInfo_Tx proto.InternalMessageInfo

func (m *Tx) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Tx) GetBundleHash() []byte {
	if m != nil {
		return m.BundleHash
	}
	return nil
}

func (m *Tx) GetRefTx() []byte {
	if m != nil {
		return m.RefTx
	}
	return nil
}

func (m *Tx) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

func (m *Tx) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *Tx) GetTag() []byte {
	if m != nil {
		return m.Tag
	}
	return nil
}

type WeakReq struct {
	Hash                 []byte   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Total                uint32   `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WeakReq) Reset()         { *m = WeakReq{} }
func (m *WeakReq) String() string { return proto.CompactTextString(m) }
func (*WeakReq) ProtoMessage()    {}
func (*WeakReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{4}
}

func (m *WeakReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WeakReq.Unmarshal(m, b)
}
func (m *WeakReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WeakReq.Marshal(b, m, deterministic)
}
func (m *WeakReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WeakReq.Merge(m, src)
}
func (m *WeakReq) XXX_Size() int {
	return xxx_messageInfo_WeakReq.Size(m)
}
func (m *WeakReq) XXX_DiscardUnknown() {
	xxx_messageInfo_WeakReq.DiscardUnknown(m)
}

var xxx_messageInfo_WeakReq proto.InternalMessageInfo

func (m *WeakReq) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *WeakReq) GetTotal() uint32 {
	if m != nil {
		return m.Total
	}
	return 0
}

type Rep struct {
	Hash                 RepType  `protobuf:"varint,1,opt,name=hash,proto3,enum=msg.RepType" json:"hash,omitempty"`
	Addr                 []byte   `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"`
	RepReview            *Review  `protobuf:"bytes,3,opt,name=rep_review,json=repReview,proto3" json:"rep_review,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Rep) Reset()         { *m = Rep{} }
func (m *Rep) String() string { return proto.CompactTextString(m) }
func (*Rep) ProtoMessage()    {}
func (*Rep) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{5}
}

func (m *Rep) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Rep.Unmarshal(m, b)
}
func (m *Rep) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Rep.Marshal(b, m, deterministic)
}
func (m *Rep) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Rep.Merge(m, src)
}
func (m *Rep) XXX_Size() int {
	return xxx_messageInfo_Rep.Size(m)
}
func (m *Rep) XXX_DiscardUnknown() {
	xxx_messageInfo_Rep.DiscardUnknown(m)
}

var xxx_messageInfo_Rep proto.InternalMessageInfo

func (m *Rep) GetHash() RepType {
	if m != nil {
		return m.Hash
	}
	return Rep_POPR
}

func (m *Rep) GetAddr() []byte {
	if m != nil {
		return m.Addr
	}
	return nil
}

func (m *Rep) GetRepReview() *Review {
	if m != nil {
		return m.RepReview
	}
	return nil
}

type Review struct {
	Rate                 uint32   `protobuf:"varint,1,opt,name=rate,proto3" json:"rate,omitempty"`
	Confidence           uint32   `protobuf:"varint,2,opt,name=confidence,proto3" json:"confidence,omitempty"`
	IpfsReview           []byte   `protobuf:"bytes,3,opt,name=ipfs_review,json=ipfsReview,proto3" json:"ipfs_review,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Review) Reset()         { *m = Review{} }
func (m *Review) String() string { return proto.CompactTextString(m) }
func (*Review) ProtoMessage()    {}
func (*Review) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{6}
}

func (m *Review) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Review.Unmarshal(m, b)
}
func (m *Review) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Review.Marshal(b, m, deterministic)
}
func (m *Review) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Review.Merge(m, src)
}
func (m *Review) XXX_Size() int {
	return xxx_messageInfo_Review.Size(m)
}
func (m *Review) XXX_DiscardUnknown() {
	xxx_messageInfo_Review.DiscardUnknown(m)
}

var xxx_messageInfo_Review proto.InternalMessageInfo

func (m *Review) GetRate() uint32 {
	if m != nil {
		return m.Rate
	}
	return 0
}

func (m *Review) GetConfidence() uint32 {
	if m != nil {
		return m.Confidence
	}
	return 0
}

func (m *Review) GetIpfsReview() []byte {
	if m != nil {
		return m.IpfsReview
	}
	return nil
}

type Initial struct {
	POBurn               *Bundle  `protobuf:"bytes,1,opt,name=p_o_burn,json=pOBurn,proto3" json:"p_o_burn,omitempty"`
	Service              []byte   `protobuf:"bytes,2,opt,name=service,proto3" json:"service,omitempty"`
	OwnerAddr            []byte   `protobuf:"bytes,3,opt,name=owner_addr,json=ownerAddr,proto3" json:"owner_addr,omitempty"`
	IpfsDetail           []byte   `protobuf:"bytes,4,opt,name=ipfs_detail,json=ipfsDetail,proto3" json:"ipfs_detail,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Initial) Reset()         { *m = Initial{} }
func (m *Initial) String() string { return proto.CompactTextString(m) }
func (*Initial) ProtoMessage()    {}
func (*Initial) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{7}
}

func (m *Initial) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Initial.Unmarshal(m, b)
}
func (m *Initial) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Initial.Marshal(b, m, deterministic)
}
func (m *Initial) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Initial.Merge(m, src)
}
func (m *Initial) XXX_Size() int {
	return xxx_messageInfo_Initial.Size(m)
}
func (m *Initial) XXX_DiscardUnknown() {
	xxx_messageInfo_Initial.DiscardUnknown(m)
}

var xxx_messageInfo_Initial proto.InternalMessageInfo

func (m *Initial) GetPOBurn() *Bundle {
	if m != nil {
		return m.POBurn
	}
	return nil
}

func (m *Initial) GetService() []byte {
	if m != nil {
		return m.Service
	}
	return nil
}

func (m *Initial) GetOwnerAddr() []byte {
	if m != nil {
		return m.OwnerAddr
	}
	return nil
}

func (m *Initial) GetIpfsDetail() []byte {
	if m != nil {
		return m.IpfsDetail
	}
	return nil
}

type SanityCheck struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SanityCheck) Reset()         { *m = SanityCheck{} }
func (m *SanityCheck) String() string { return proto.CompactTextString(m) }
func (*SanityCheck) ProtoMessage()    {}
func (*SanityCheck) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fc5e78786d5f342, []int{8}
}

func (m *SanityCheck) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SanityCheck.Unmarshal(m, b)
}
func (m *SanityCheck) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SanityCheck.Marshal(b, m, deterministic)
}
func (m *SanityCheck) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SanityCheck.Merge(m, src)
}
func (m *SanityCheck) XXX_Size() int {
	return xxx_messageInfo_SanityCheck.Size(m)
}
func (m *SanityCheck) XXX_DiscardUnknown() {
	xxx_messageInfo_SanityCheck.DiscardUnknown(m)
}

var xxx_messageInfo_SanityCheck proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("msg.PacketType", PacketType_name, PacketType_value)
	proto.RegisterEnum("msg.BundleType", BundleType_name, BundleType_value)
	proto.RegisterEnum("msg.RepType", RepType_name, RepType_value)
	proto.RegisterType((*Packet)(nil), "msg.packet")
	proto.RegisterType((*Block)(nil), "msg.block")
	proto.RegisterType((*Bundle)(nil), "msg.bundle")
	proto.RegisterType((*Tx)(nil), "msg.tx")
	proto.RegisterType((*WeakReq)(nil), "msg.weak_req")
	proto.RegisterType((*Rep)(nil), "msg.rep")
	proto.RegisterType((*Review)(nil), "msg.review")
	proto.RegisterType((*Initial)(nil), "msg.initial")
	proto.RegisterType((*SanityCheck)(nil), "msg.sanity_check")
}

func init() { proto.RegisterFile("Messages/Messages.proto", fileDescriptor_3fc5e78786d5f342) }

var fileDescriptor_3fc5e78786d5f342 = []byte{
	// 848 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x54, 0x51, 0x6f, 0xe2, 0x46,
	0x10, 0x66, 0xb1, 0x01, 0x33, 0x40, 0xea, 0xae, 0x90, 0x6a, 0x9d, 0xd4, 0x3b, 0x6a, 0xf5, 0x24,
	0x74, 0xa7, 0x23, 0xbd, 0xdc, 0x3d, 0xf4, 0x15, 0x12, 0x57, 0xa0, 0xa4, 0x49, 0xba, 0xe5, 0x1a,
	0xf5, 0xa1, 0xb2, 0x16, 0x58, 0x88, 0x75, 0x60, 0xfb, 0xd6, 0x0b, 0x49, 0xde, 0xfa, 0xdc, 0xfe,
	0x87, 0xfe, 0xa6, 0xaa, 0xbf, 0xa8, 0xda, 0x59, 0xdb, 0x71, 0x9b, 0x7b, 0x9b, 0xf9, 0xe6, 0xf3,
	0xcc, 0xb7, 0xb3, 0xdf, 0x1a, 0xbe, 0xfa, 0x51, 0x64, 0x19, 0xdf, 0x88, 0xec, 0xb8, 0x08, 0x46,
	0xa9, 0x4c, 0x54, 0x42, 0xad, 0x5d, 0xb6, 0x79, 0xf6, 0x62, 0x93, 0x24, 0x9b, 0xad, 0x38, 0x46,
	0x68, 0xb1, 0x5f, 0x1f, 0xab, 0x68, 0x27, 0x32, 0xc5, 0x77, 0xa9, 0x61, 0xf9, 0xff, 0xd8, 0xd0,
	0x4c, 0xf9, 0xf2, 0xa3, 0x50, 0xf4, 0x2d, 0x74, 0x4c, 0x14, 0xaa, 0x87, 0x54, 0x78, 0x64, 0x40,
	0x86, 0x47, 0x27, 0xee, 0x68, 0x97, 0x6d, 0x46, 0x06, 0x1f, 0x69, 0x9c, 0x81, 0x49, 0xe6, 0x0f,
	0xa9, 0xa0, 0xdf, 0x43, 0xbb, 0x6c, 0xe8, 0xd5, 0x07, 0x64, 0xd8, 0x39, 0x79, 0x36, 0x32, 0x23,
	0x47, 0xc5, 0xc8, 0xd1, 0xbc, 0x60, 0xb0, 0x47, 0x32, 0xa5, 0x60, 0xa7, 0x52, 0x1c, 0x3c, 0x6b,
	0x40, 0x86, 0x5d, 0x86, 0x31, 0xed, 0x43, 0x23, 0x4e, 0xe2, 0xa5, 0xf0, 0x6c, 0x04, 0x4d, 0xa2,
	0x99, 0xab, 0x68, 0xbd, 0xf6, 0x1a, 0x03, 0x32, 0xec, 0x31, 0x8c, 0x35, 0x76, 0xcb, 0xb3, 0x5b,
	0xaf, 0x69, 0xbe, 0xd6, 0x31, 0xfd, 0x0e, 0xfa, 0xcb, 0xbd, 0x94, 0x22, 0x56, 0xe1, 0x62, 0x9b,
	0x2c, 0x3f, 0x86, 0xf1, 0x7e, 0xb7, 0x10, 0xd2, 0x6b, 0x0d, 0xc8, 0xd0, 0x66, 0x34, 0xaf, 0x4d,
	0x74, 0xe9, 0x12, 0x2b, 0xba, 0x0b, 0x5f, 0xad, 0xa4, 0xd7, 0x33, 0x5d, 0x74, 0xac, 0xb1, 0x2c,
	0xda, 0xc4, 0xde, 0x91, 0xc1, 0x74, 0x4c, 0x5f, 0x41, 0x1b, 0x3b, 0x9e, 0x71, 0xc5, 0x3d, 0x07,
	0x4f, 0x09, 0xb8, 0x16, 0x44, 0xa7, 0x35, 0xf6, 0x58, 0xa6, 0x6f, 0x00, 0x16, 0xfb, 0x78, 0xb5,
	0x15, 0x48, 0x6e, 0x23, 0xb9, 0x63, 0xc8, 0x08, 0x4f, 0x6b, 0xac, 0x42, 0xa0, 0xdf, 0x42, 0x4b,
	0x8a, 0x14, 0xb9, 0x80, 0x5c, 0x07, 0xb9, 0x52, 0xa4, 0xd3, 0x1a, 0x2b, 0x4a, 0xf4, 0x35, 0x38,
	0x77, 0x82, 0x9b, 0xf9, 0x1d, 0xa4, 0xf5, 0x90, 0xa6, 0xc1, 0x50, 0x8a, 0x4f, 0xd3, 0x1a, 0x2b,
	0x09, 0xf4, 0x1d, 0x40, 0xc6, 0xe3, 0x48, 0x3d, 0x20, 0xbd, 0x8b, 0xf4, 0x2f, 0x91, 0x6e, 0xe0,
	0x70, 0x79, 0x2b, 0x50, 0x75, 0x85, 0xe6, 0x9f, 0x81, 0xad, 0x2f, 0x97, 0x02, 0x34, 0x27, 0x1f,
	0x2e, 0xcf, 0x2e, 0x02, 0xb7, 0x46, 0xdb, 0xd0, 0x98, 0x5c, 0x5c, 0x9d, 0x9e, 0xbb, 0x84, 0xb6,
	0xc0, 0x62, 0xc1, 0xb5, 0x5b, 0xa7, 0x1d, 0x68, 0xdd, 0x04, 0xe3, 0x73, 0x16, 0xfc, 0xe4, 0x5a,
	0xf4, 0x0b, 0xe8, 0xfc, 0x3c, 0xbe, 0x9c, 0xcd, 0x7f, 0x3d, 0x9d, 0x06, 0xa7, 0xe7, 0xae, 0x3d,
	0x69, 0x82, 0xbd, 0xd2, 0xdd, 0x7e, 0x27, 0xd0, 0xc0, 0x95, 0xd0, 0x6f, 0xc0, 0x96, 0xe2, 0x53,
	0xe6, 0x91, 0x81, 0xf5, 0x44, 0x35, 0xc3, 0x12, 0x7d, 0x09, 0x2d, 0xb3, 0x90, 0xcc, 0xab, 0x23,
	0xab, 0xba, 0x2e, 0x56, 0xd4, 0xe8, 0x1b, 0x70, 0x50, 0x6f, 0x24, 0x32, 0xcf, 0x42, 0xde, 0xd3,
	0x43, 0xb1, 0x92, 0xe2, 0xff, 0x4d, 0xa0, 0x69, 0x3e, 0xd5, 0xbe, 0x36, 0xd1, 0x53, 0x5f, 0x1b,
	0x3c, 0xf7, 0xb5, 0x49, 0xd0, 0xd7, 0x85, 0xbf, 0xea, 0x15, 0x7f, 0xbd, 0x86, 0xae, 0x92, 0x3c,
	0xce, 0xf8, 0x52, 0x45, 0x49, 0x5c, 0x88, 0x68, 0x61, 0x1f, 0x75, 0xcf, 0xfe, 0x53, 0xa4, 0x1e,
	0xb4, 0x0e, 0x42, 0x46, 0xeb, 0x87, 0xb7, 0xb9, 0x99, 0x8b, 0xf4, 0xb1, 0x72, 0x82, 0x8e, 0x2e,
	0x2b, 0x27, 0xfe, 0xf3, 0xfc, 0x0e, 0xba, 0xe0, 0x5c, 0x5f, 0xdd, 0x04, 0xec, 0x87, 0x0f, 0x17,
	0x6e, 0x8d, 0x3a, 0x60, 0xeb, 0x8d, 0xbb, 0xc4, 0xff, 0x93, 0x40, 0x5d, 0xdd, 0x97, 0xda, 0x48,
	0x45, 0xdb, 0x8b, 0xf2, 0x88, 0x95, 0x67, 0x91, 0x1f, 0x68, 0xaa, 0x09, 0x7d, 0x68, 0x48, 0xb1,
	0x9e, 0xdf, 0xe7, 0x27, 0x32, 0x49, 0x69, 0x76, 0xab, 0x62, 0xf6, 0x3e, 0x34, 0x0e, 0x7c, 0xbb,
	0x37, 0x8f, 0xd0, 0x62, 0x26, 0xa1, 0x2e, 0x58, 0x8a, 0x6f, 0x72, 0xc5, 0x3a, 0xf4, 0xdf, 0x1b,
	0x4f, 0xea, 0x8b, 0xfc, 0xac, 0xa4, 0x3e, 0x34, 0x54, 0xa2, 0xf8, 0x16, 0x27, 0xf6, 0x98, 0x49,
	0xfc, 0xbf, 0x08, 0x58, 0x52, 0xa4, 0xda, 0x17, 0xe5, 0x17, 0x47, 0xb9, 0x2f, 0xa4, 0x48, 0xcd,
	0x4d, 0x98, 0x06, 0xc5, 0xeb, 0xac, 0x57, 0x5e, 0xe7, 0x2b, 0x00, 0x29, 0xd2, 0x50, 0x8a, 0x43,
	0x24, 0xee, 0x50, 0x76, 0x61, 0x17, 0x03, 0xb1, 0xb6, 0x14, 0x29, 0xc3, 0xd0, 0x7f, 0x9f, 0xaf,
	0xd3, 0x01, 0xfb, 0xfa, 0xea, 0x9a, 0x19, 0x43, 0xeb, 0x68, 0xec, 0x12, 0xed, 0x73, 0x16, 0xfc,
	0x32, 0x0b, 0x6e, 0x8c, 0xa7, 0x67, 0x97, 0xb3, 0xf9, 0x6c, 0x7c, 0xe1, 0x5a, 0xfe, 0x6f, 0xd0,
	0x34, 0xad, 0xf4, 0x7c, 0xc9, 0x95, 0xf1, 0x4b, 0x8f, 0x61, 0x4c, 0x9f, 0x03, 0x2c, 0x93, 0x78,
	0x1d, 0xad, 0x84, 0xfe, 0x4d, 0x99, 0x93, 0x55, 0x10, 0x7d, 0x0f, 0x51, 0xba, 0xce, 0xaa, 0x02,
	0xbb, 0x0c, 0x34, 0x94, 0x8b, 0xfa, 0x83, 0x40, 0x2b, 0xd2, 0x16, 0xe5, 0x5b, 0xfa, 0x12, 0x9c,
	0x34, 0x4c, 0xc2, 0xc5, 0x5e, 0xc6, 0x38, 0xe4, 0x7f, 0xce, 0x6f, 0xa6, 0x57, 0x93, 0xbd, 0x8c,
	0xb5, 0x61, 0x32, 0x21, 0x0f, 0x51, 0x3e, 0xb0, 0xcb, 0x8a, 0x94, 0x7e, 0x0d, 0x90, 0xdc, 0xc5,
	0x42, 0x86, 0xb8, 0x27, 0x33, 0xac, 0x8d, 0xc8, 0x58, 0x2f, 0xab, 0x10, 0xb3, 0x12, 0x8a, 0x47,
	0xdb, 0xdc, 0x87, 0x28, 0xe6, 0x0c, 0x11, 0xff, 0x08, 0xba, 0xd5, 0xd7, 0xb3, 0x68, 0xe2, 0x2f,
	0xfb, 0xdd, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x37, 0x7c, 0xa8, 0xec, 0x53, 0x06, 0x00, 0x00,
}

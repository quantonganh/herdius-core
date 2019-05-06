// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hbi/protobuf/service.proto

package protobuf

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Timestamp struct {
	Seconds              int64    `protobuf:"varint,1,opt,name=seconds,proto3" json:"seconds,omitempty"`
	Nanos                int64    `protobuf:"varint,2,opt,name=nanos,proto3" json:"nanos,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Timestamp) Reset()         { *m = Timestamp{} }
func (m *Timestamp) String() string { return proto.CompactTextString(m) }
func (*Timestamp) ProtoMessage()    {}
func (*Timestamp) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{0}
}

func (m *Timestamp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Timestamp.Unmarshal(m, b)
}
func (m *Timestamp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Timestamp.Marshal(b, m, deterministic)
}
func (m *Timestamp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Timestamp.Merge(m, src)
}
func (m *Timestamp) XXX_Size() int {
	return xxx_messageInfo_Timestamp.Size(m)
}
func (m *Timestamp) XXX_DiscardUnknown() {
	xxx_messageInfo_Timestamp.DiscardUnknown(m)
}

var xxx_messageInfo_Timestamp proto.InternalMessageInfo

func (m *Timestamp) GetSeconds() int64 {
	if m != nil {
		return m.Seconds
	}
	return 0
}

func (m *Timestamp) GetNanos() int64 {
	if m != nil {
		return m.Nanos
	}
	return 0
}

type BlockHeightRequest struct {
	BlockHeight          int64    `protobuf:"varint,1,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockHeightRequest) Reset()         { *m = BlockHeightRequest{} }
func (m *BlockHeightRequest) String() string { return proto.CompactTextString(m) }
func (*BlockHeightRequest) ProtoMessage()    {}
func (*BlockHeightRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{1}
}

func (m *BlockHeightRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockHeightRequest.Unmarshal(m, b)
}
func (m *BlockHeightRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockHeightRequest.Marshal(b, m, deterministic)
}
func (m *BlockHeightRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockHeightRequest.Merge(m, src)
}
func (m *BlockHeightRequest) XXX_Size() int {
	return xxx_messageInfo_BlockHeightRequest.Size(m)
}
func (m *BlockHeightRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockHeightRequest.DiscardUnknown(m)
}

var xxx_messageInfo_BlockHeightRequest proto.InternalMessageInfo

func (m *BlockHeightRequest) GetBlockHeight() int64 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

type BlockResponse struct {
	BlockHeight int64 `protobuf:"varint,1,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	// Time of block intialization
	Time     *Timestamp `protobuf:"bytes,2,opt,name=time,proto3" json:"time,omitempty"`
	TotalTxs uint64     `protobuf:"varint,3,opt,name=total_txs,json=totalTxs,proto3" json:"total_txs,omitempty"`
	// Supervisor herdius token address who created the block
	SupervisorAddress    string   `protobuf:"bytes,4,opt,name=supervisor_address,json=supervisorAddress,proto3" json:"supervisor_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockResponse) Reset()         { *m = BlockResponse{} }
func (m *BlockResponse) String() string { return proto.CompactTextString(m) }
func (*BlockResponse) ProtoMessage()    {}
func (*BlockResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{2}
}

func (m *BlockResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockResponse.Unmarshal(m, b)
}
func (m *BlockResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockResponse.Marshal(b, m, deterministic)
}
func (m *BlockResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockResponse.Merge(m, src)
}
func (m *BlockResponse) XXX_Size() int {
	return xxx_messageInfo_BlockResponse.Size(m)
}
func (m *BlockResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockResponse.DiscardUnknown(m)
}

var xxx_messageInfo_BlockResponse proto.InternalMessageInfo

func (m *BlockResponse) GetBlockHeight() int64 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *BlockResponse) GetTime() *Timestamp {
	if m != nil {
		return m.Time
	}
	return nil
}

func (m *BlockResponse) GetTotalTxs() uint64 {
	if m != nil {
		return m.TotalTxs
	}
	return 0
}

func (m *BlockResponse) GetSupervisorAddress() string {
	if m != nil {
		return m.SupervisorAddress
	}
	return ""
}

type AccountRequest struct {
	Address              string   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccountRequest) Reset()         { *m = AccountRequest{} }
func (m *AccountRequest) String() string { return proto.CompactTextString(m) }
func (*AccountRequest) ProtoMessage()    {}
func (*AccountRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{3}
}

func (m *AccountRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountRequest.Unmarshal(m, b)
}
func (m *AccountRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountRequest.Marshal(b, m, deterministic)
}
func (m *AccountRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountRequest.Merge(m, src)
}
func (m *AccountRequest) XXX_Size() int {
	return xxx_messageInfo_AccountRequest.Size(m)
}
func (m *AccountRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AccountRequest proto.InternalMessageInfo

func (m *AccountRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type AccountResponse struct {
	Address              string            `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Nonce                uint64            `protobuf:"varint,2,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Balance              uint64            `protobuf:"varint,3,opt,name=balance,proto3" json:"balance,omitempty"`
	StorageRoot          string            `protobuf:"bytes,4,opt,name=storage_root,json=storageRoot,proto3" json:"storage_root,omitempty"`
	PublicKey            string            `protobuf:"bytes,5,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	Balances             map[string]uint64 `protobuf:"bytes,6,rep,name=balances,proto3" json:"balances,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *AccountResponse) Reset()         { *m = AccountResponse{} }
func (m *AccountResponse) String() string { return proto.CompactTextString(m) }
func (*AccountResponse) ProtoMessage()    {}
func (*AccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{4}
}

func (m *AccountResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountResponse.Unmarshal(m, b)
}
func (m *AccountResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountResponse.Marshal(b, m, deterministic)
}
func (m *AccountResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountResponse.Merge(m, src)
}
func (m *AccountResponse) XXX_Size() int {
	return xxx_messageInfo_AccountResponse.Size(m)
}
func (m *AccountResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AccountResponse proto.InternalMessageInfo

func (m *AccountResponse) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *AccountResponse) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *AccountResponse) GetBalance() uint64 {
	if m != nil {
		return m.Balance
	}
	return 0
}

func (m *AccountResponse) GetStorageRoot() string {
	if m != nil {
		return m.StorageRoot
	}
	return ""
}

func (m *AccountResponse) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

func (m *AccountResponse) GetBalances() map[string]uint64 {
	if m != nil {
		return m.Balances
	}
	return nil
}

type Asset struct {
	Category             string   `protobuf:"bytes,1,opt,name=category,proto3" json:"category,omitempty"`
	Symbol               string   `protobuf:"bytes,2,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Network              string   `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Value                uint64   `protobuf:"varint,4,opt,name=value,proto3" json:"value,omitempty"`
	Fee                  uint64   `protobuf:"varint,5,opt,name=fee,proto3" json:"fee,omitempty"`
	Nonce                uint64   `protobuf:"varint,6,opt,name=nonce,proto3" json:"nonce,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Asset) Reset()         { *m = Asset{} }
func (m *Asset) String() string { return proto.CompactTextString(m) }
func (*Asset) ProtoMessage()    {}
func (*Asset) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{5}
}

func (m *Asset) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Asset.Unmarshal(m, b)
}
func (m *Asset) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Asset.Marshal(b, m, deterministic)
}
func (m *Asset) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Asset.Merge(m, src)
}
func (m *Asset) XXX_Size() int {
	return xxx_messageInfo_Asset.Size(m)
}
func (m *Asset) XXX_DiscardUnknown() {
	xxx_messageInfo_Asset.DiscardUnknown(m)
}

var xxx_messageInfo_Asset proto.InternalMessageInfo

func (m *Asset) GetCategory() string {
	if m != nil {
		return m.Category
	}
	return ""
}

func (m *Asset) GetSymbol() string {
	if m != nil {
		return m.Symbol
	}
	return ""
}

func (m *Asset) GetNetwork() string {
	if m != nil {
		return m.Network
	}
	return ""
}

func (m *Asset) GetValue() uint64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *Asset) GetFee() uint64 {
	if m != nil {
		return m.Fee
	}
	return 0
}

func (m *Asset) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

type Tx struct {
	SenderAddress   string `protobuf:"bytes,1,opt,name=sender_address,json=senderAddress,proto3" json:"sender_address,omitempty"`
	SenderPubkey    string `protobuf:"bytes,2,opt,name=sender_pubkey,json=senderPubkey,proto3" json:"sender_pubkey,omitempty"`
	RecieverAddress string `protobuf:"bytes,3,opt,name=reciever_address,json=recieverAddress,proto3" json:"reciever_address,omitempty"`
	Asset           *Asset `protobuf:"bytes,4,opt,name=asset,proto3" json:"asset,omitempty"`
	Message         string `protobuf:"bytes,5,opt,name=message,proto3" json:"message,omitempty"`
	Sign            string `protobuf:"bytes,6,opt,name=sign,proto3" json:"sign,omitempty"`
	// type will check if tx is of type Account Registeration or Value Transfer
	Type                 string   `protobuf:"bytes,7,opt,name=type,proto3" json:"type,omitempty"`
	Status               string   `protobuf:"bytes,8,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Tx) Reset()         { *m = Tx{} }
func (m *Tx) String() string { return proto.CompactTextString(m) }
func (*Tx) ProtoMessage()    {}
func (*Tx) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{6}
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

func (m *Tx) GetSenderAddress() string {
	if m != nil {
		return m.SenderAddress
	}
	return ""
}

func (m *Tx) GetSenderPubkey() string {
	if m != nil {
		return m.SenderPubkey
	}
	return ""
}

func (m *Tx) GetRecieverAddress() string {
	if m != nil {
		return m.RecieverAddress
	}
	return ""
}

func (m *Tx) GetAsset() *Asset {
	if m != nil {
		return m.Asset
	}
	return nil
}

func (m *Tx) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *Tx) GetSign() string {
	if m != nil {
		return m.Sign
	}
	return ""
}

func (m *Tx) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Tx) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

type TxRequest struct {
	Tx                   *Tx      `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TxRequest) Reset()         { *m = TxRequest{} }
func (m *TxRequest) String() string { return proto.CompactTextString(m) }
func (*TxRequest) ProtoMessage()    {}
func (*TxRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{7}
}

func (m *TxRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TxRequest.Unmarshal(m, b)
}
func (m *TxRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TxRequest.Marshal(b, m, deterministic)
}
func (m *TxRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TxRequest.Merge(m, src)
}
func (m *TxRequest) XXX_Size() int {
	return xxx_messageInfo_TxRequest.Size(m)
}
func (m *TxRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TxRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TxRequest proto.InternalMessageInfo

func (m *TxRequest) GetTx() *Tx {
	if m != nil {
		return m.Tx
	}
	return nil
}

type TxResponse struct {
	TxId                 string   `protobuf:"bytes,1,opt,name=tx_id,json=txId,proto3" json:"tx_id,omitempty"`
	Pending              int64    `protobuf:"varint,2,opt,name=pending,proto3" json:"pending,omitempty"`
	Queued               int64    `protobuf:"varint,3,opt,name=queued,proto3" json:"queued,omitempty"`
	Status               string   `protobuf:"bytes,4,opt,name=status,proto3" json:"status,omitempty"`
	Message              string   `protobuf:"bytes,5,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TxResponse) Reset()         { *m = TxResponse{} }
func (m *TxResponse) String() string { return proto.CompactTextString(m) }
func (*TxResponse) ProtoMessage()    {}
func (*TxResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{8}
}

func (m *TxResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TxResponse.Unmarshal(m, b)
}
func (m *TxResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TxResponse.Marshal(b, m, deterministic)
}
func (m *TxResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TxResponse.Merge(m, src)
}
func (m *TxResponse) XXX_Size() int {
	return xxx_messageInfo_TxResponse.Size(m)
}
func (m *TxResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TxResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TxResponse proto.InternalMessageInfo

func (m *TxResponse) GetTxId() string {
	if m != nil {
		return m.TxId
	}
	return ""
}

func (m *TxResponse) GetPending() int64 {
	if m != nil {
		return m.Pending
	}
	return 0
}

func (m *TxResponse) GetQueued() int64 {
	if m != nil {
		return m.Queued
	}
	return 0
}

func (m *TxResponse) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *TxResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type AccountRegisterRequest struct {
	SenderPubkey         string   `protobuf:"bytes,1,opt,name=sender_pubkey,json=senderPubkey,proto3" json:"sender_pubkey,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccountRegisterRequest) Reset()         { *m = AccountRegisterRequest{} }
func (m *AccountRegisterRequest) String() string { return proto.CompactTextString(m) }
func (*AccountRegisterRequest) ProtoMessage()    {}
func (*AccountRegisterRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{9}
}

func (m *AccountRegisterRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountRegisterRequest.Unmarshal(m, b)
}
func (m *AccountRegisterRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountRegisterRequest.Marshal(b, m, deterministic)
}
func (m *AccountRegisterRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountRegisterRequest.Merge(m, src)
}
func (m *AccountRegisterRequest) XXX_Size() int {
	return xxx_messageInfo_AccountRegisterRequest.Size(m)
}
func (m *AccountRegisterRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountRegisterRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AccountRegisterRequest proto.InternalMessageInfo

func (m *AccountRegisterRequest) GetSenderPubkey() string {
	if m != nil {
		return m.SenderPubkey
	}
	return ""
}

// Send request to retrieve transaction committed in herdius blockchain
type TxDetailRequest struct {
	TxId                 string   `protobuf:"bytes,1,opt,name=tx_id,json=txId,proto3" json:"tx_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TxDetailRequest) Reset()         { *m = TxDetailRequest{} }
func (m *TxDetailRequest) String() string { return proto.CompactTextString(m) }
func (*TxDetailRequest) ProtoMessage()    {}
func (*TxDetailRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{10}
}

func (m *TxDetailRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TxDetailRequest.Unmarshal(m, b)
}
func (m *TxDetailRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TxDetailRequest.Marshal(b, m, deterministic)
}
func (m *TxDetailRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TxDetailRequest.Merge(m, src)
}
func (m *TxDetailRequest) XXX_Size() int {
	return xxx_messageInfo_TxDetailRequest.Size(m)
}
func (m *TxDetailRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TxDetailRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TxDetailRequest proto.InternalMessageInfo

func (m *TxDetailRequest) GetTxId() string {
	if m != nil {
		return m.TxId
	}
	return ""
}

// Transaction detail response from herdius blockchain
type TxDetailResponse struct {
	TxId                 string     `protobuf:"bytes,1,opt,name=tx_id,json=txId,proto3" json:"tx_id,omitempty"`
	Tx                   *Tx        `protobuf:"bytes,2,opt,name=tx,proto3" json:"tx,omitempty"`
	CreationDt           *Timestamp `protobuf:"bytes,3,opt,name=creationDt,proto3" json:"creationDt,omitempty"`
	BlockId              uint64     `protobuf:"varint,4,opt,name=block_id,json=blockId,proto3" json:"block_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *TxDetailResponse) Reset()         { *m = TxDetailResponse{} }
func (m *TxDetailResponse) String() string { return proto.CompactTextString(m) }
func (*TxDetailResponse) ProtoMessage()    {}
func (*TxDetailResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{11}
}

func (m *TxDetailResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TxDetailResponse.Unmarshal(m, b)
}
func (m *TxDetailResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TxDetailResponse.Marshal(b, m, deterministic)
}
func (m *TxDetailResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TxDetailResponse.Merge(m, src)
}
func (m *TxDetailResponse) XXX_Size() int {
	return xxx_messageInfo_TxDetailResponse.Size(m)
}
func (m *TxDetailResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TxDetailResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TxDetailResponse proto.InternalMessageInfo

func (m *TxDetailResponse) GetTxId() string {
	if m != nil {
		return m.TxId
	}
	return ""
}

func (m *TxDetailResponse) GetTx() *Tx {
	if m != nil {
		return m.Tx
	}
	return nil
}

func (m *TxDetailResponse) GetCreationDt() *Timestamp {
	if m != nil {
		return m.CreationDt
	}
	return nil
}

func (m *TxDetailResponse) GetBlockId() uint64 {
	if m != nil {
		return m.BlockId
	}
	return 0
}

type Transaction struct {
	Senderpubkey         []byte   `protobuf:"bytes,1,opt,name=senderpubkey,proto3" json:"senderpubkey,omitempty"`
	Signature            string   `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	Recaddress           string   `protobuf:"bytes,3,opt,name=recaddress,proto3" json:"recaddress,omitempty"`
	Asset                *Asset   `protobuf:"bytes,4,opt,name=asset,proto3" json:"asset,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Transaction) Reset()         { *m = Transaction{} }
func (m *Transaction) String() string { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()    {}
func (*Transaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{12}
}

func (m *Transaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction.Unmarshal(m, b)
}
func (m *Transaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction.Marshal(b, m, deterministic)
}
func (m *Transaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction.Merge(m, src)
}
func (m *Transaction) XXX_Size() int {
	return xxx_messageInfo_Transaction.Size(m)
}
func (m *Transaction) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction proto.InternalMessageInfo

func (m *Transaction) GetSenderpubkey() []byte {
	if m != nil {
		return m.Senderpubkey
	}
	return nil
}

func (m *Transaction) GetSignature() string {
	if m != nil {
		return m.Signature
	}
	return ""
}

func (m *Transaction) GetRecaddress() string {
	if m != nil {
		return m.Recaddress
	}
	return ""
}

func (m *Transaction) GetAsset() *Asset {
	if m != nil {
		return m.Asset
	}
	return nil
}

type TransactionRequest struct {
	Tx                   *Transaction `protobuf:"bytes,1,opt,name=Tx,proto3" json:"Tx,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *TransactionRequest) Reset()         { *m = TransactionRequest{} }
func (m *TransactionRequest) String() string { return proto.CompactTextString(m) }
func (*TransactionRequest) ProtoMessage()    {}
func (*TransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{13}
}

func (m *TransactionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TransactionRequest.Unmarshal(m, b)
}
func (m *TransactionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TransactionRequest.Marshal(b, m, deterministic)
}
func (m *TransactionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TransactionRequest.Merge(m, src)
}
func (m *TransactionRequest) XXX_Size() int {
	return xxx_messageInfo_TransactionRequest.Size(m)
}
func (m *TransactionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TransactionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TransactionRequest proto.InternalMessageInfo

func (m *TransactionRequest) GetTx() *Transaction {
	if m != nil {
		return m.Tx
	}
	return nil
}

type TransactionResponse struct {
	TxId                 string   `protobuf:"bytes,1,opt,name=tx_id,json=txId,proto3" json:"tx_id,omitempty"`
	Pending              int64    `protobuf:"varint,2,opt,name=pending,proto3" json:"pending,omitempty"`
	Queued               int64    `protobuf:"varint,3,opt,name=queued,proto3" json:"queued,omitempty"`
	Status               string   `protobuf:"bytes,4,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TransactionResponse) Reset()         { *m = TransactionResponse{} }
func (m *TransactionResponse) String() string { return proto.CompactTextString(m) }
func (*TransactionResponse) ProtoMessage()    {}
func (*TransactionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{14}
}

func (m *TransactionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TransactionResponse.Unmarshal(m, b)
}
func (m *TransactionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TransactionResponse.Marshal(b, m, deterministic)
}
func (m *TransactionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TransactionResponse.Merge(m, src)
}
func (m *TransactionResponse) XXX_Size() int {
	return xxx_messageInfo_TransactionResponse.Size(m)
}
func (m *TransactionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TransactionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TransactionResponse proto.InternalMessageInfo

func (m *TransactionResponse) GetTxId() string {
	if m != nil {
		return m.TxId
	}
	return ""
}

func (m *TransactionResponse) GetPending() int64 {
	if m != nil {
		return m.Pending
	}
	return 0
}

func (m *TransactionResponse) GetQueued() int64 {
	if m != nil {
		return m.Queued
	}
	return 0
}

func (m *TransactionResponse) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

// Send request to retrieve all transactions of an address committed in herdius blockchain
type TxsByAddressRequest struct {
	Address              string   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TxsByAddressRequest) Reset()         { *m = TxsByAddressRequest{} }
func (m *TxsByAddressRequest) String() string { return proto.CompactTextString(m) }
func (*TxsByAddressRequest) ProtoMessage()    {}
func (*TxsByAddressRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{15}
}

func (m *TxsByAddressRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TxsByAddressRequest.Unmarshal(m, b)
}
func (m *TxsByAddressRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TxsByAddressRequest.Marshal(b, m, deterministic)
}
func (m *TxsByAddressRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TxsByAddressRequest.Merge(m, src)
}
func (m *TxsByAddressRequest) XXX_Size() int {
	return xxx_messageInfo_TxsByAddressRequest.Size(m)
}
func (m *TxsByAddressRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TxsByAddressRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TxsByAddressRequest proto.InternalMessageInfo

func (m *TxsByAddressRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

// Transactions details response from herdius blockchain
type TxsResponse struct {
	Txs                  []*TxDetailResponse `protobuf:"bytes,1,rep,name=txs,proto3" json:"txs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *TxsResponse) Reset()         { *m = TxsResponse{} }
func (m *TxsResponse) String() string { return proto.CompactTextString(m) }
func (*TxsResponse) ProtoMessage()    {}
func (*TxsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{16}
}

func (m *TxsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TxsResponse.Unmarshal(m, b)
}
func (m *TxsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TxsResponse.Marshal(b, m, deterministic)
}
func (m *TxsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TxsResponse.Merge(m, src)
}
func (m *TxsResponse) XXX_Size() int {
	return xxx_messageInfo_TxsResponse.Size(m)
}
func (m *TxsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TxsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TxsResponse proto.InternalMessageInfo

func (m *TxsResponse) GetTxs() []*TxDetailResponse {
	if m != nil {
		return m.Txs
	}
	return nil
}

// Send request to retrieve all transactions of an address and asset committed in herdius blockchain
type TxsByAssetAndAddressRequest struct {
	Address              string   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Asset                string   `protobuf:"bytes,2,opt,name=asset,proto3" json:"asset,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TxsByAssetAndAddressRequest) Reset()         { *m = TxsByAssetAndAddressRequest{} }
func (m *TxsByAssetAndAddressRequest) String() string { return proto.CompactTextString(m) }
func (*TxsByAssetAndAddressRequest) ProtoMessage()    {}
func (*TxsByAssetAndAddressRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebece8ff681ed6b0, []int{17}
}

func (m *TxsByAssetAndAddressRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TxsByAssetAndAddressRequest.Unmarshal(m, b)
}
func (m *TxsByAssetAndAddressRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TxsByAssetAndAddressRequest.Marshal(b, m, deterministic)
}
func (m *TxsByAssetAndAddressRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TxsByAssetAndAddressRequest.Merge(m, src)
}
func (m *TxsByAssetAndAddressRequest) XXX_Size() int {
	return xxx_messageInfo_TxsByAssetAndAddressRequest.Size(m)
}
func (m *TxsByAssetAndAddressRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TxsByAssetAndAddressRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TxsByAssetAndAddressRequest proto.InternalMessageInfo

func (m *TxsByAssetAndAddressRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *TxsByAssetAndAddressRequest) GetAsset() string {
	if m != nil {
		return m.Asset
	}
	return ""
}

func init() {
	proto.RegisterType((*Timestamp)(nil), "protobuf.Timestamp")
	proto.RegisterType((*BlockHeightRequest)(nil), "protobuf.BlockHeightRequest")
	proto.RegisterType((*BlockResponse)(nil), "protobuf.BlockResponse")
	proto.RegisterType((*AccountRequest)(nil), "protobuf.AccountRequest")
	proto.RegisterType((*AccountResponse)(nil), "protobuf.AccountResponse")
	proto.RegisterMapType((map[string]uint64)(nil), "protobuf.AccountResponse.BalancesEntry")
	proto.RegisterType((*Asset)(nil), "protobuf.Asset")
	proto.RegisterType((*Tx)(nil), "protobuf.Tx")
	proto.RegisterType((*TxRequest)(nil), "protobuf.TxRequest")
	proto.RegisterType((*TxResponse)(nil), "protobuf.TxResponse")
	proto.RegisterType((*AccountRegisterRequest)(nil), "protobuf.AccountRegisterRequest")
	proto.RegisterType((*TxDetailRequest)(nil), "protobuf.TxDetailRequest")
	proto.RegisterType((*TxDetailResponse)(nil), "protobuf.TxDetailResponse")
	proto.RegisterType((*Transaction)(nil), "protobuf.Transaction")
	proto.RegisterType((*TransactionRequest)(nil), "protobuf.TransactionRequest")
	proto.RegisterType((*TransactionResponse)(nil), "protobuf.TransactionResponse")
	proto.RegisterType((*TxsByAddressRequest)(nil), "protobuf.TxsByAddressRequest")
	proto.RegisterType((*TxsResponse)(nil), "protobuf.TxsResponse")
	proto.RegisterType((*TxsByAssetAndAddressRequest)(nil), "protobuf.TxsByAssetAndAddressRequest")
}

func init() { proto.RegisterFile("hbi/protobuf/service.proto", fileDescriptor_ebece8ff681ed6b0) }

var fileDescriptor_ebece8ff681ed6b0 = []byte{
	// 870 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x55, 0xcf, 0x6f, 0xdc, 0x44,
	0x18, 0x95, 0x77, 0xbd, 0xc9, 0xfa, 0xdb, 0x4d, 0x93, 0x4e, 0xa0, 0x32, 0x69, 0x41, 0x61, 0x50,
	0x68, 0x8a, 0xe8, 0x46, 0x4a, 0x0f, 0x20, 0x22, 0x0e, 0x59, 0x8a, 0xd4, 0x0a, 0x21, 0x45, 0x23,
	0x9f, 0xb8, 0xac, 0xc6, 0xf6, 0xd7, 0x8d, 0x95, 0x5d, 0xcf, 0xd6, 0x33, 0x0e, 0xde, 0x3f, 0x80,
	0xbf, 0x00, 0xc1, 0x95, 0x23, 0xff, 0x24, 0x07, 0x34, 0xbf, 0xbc, 0x5e, 0xda, 0x40, 0x4e, 0xdc,
	0xe6, 0xbd, 0xf9, 0x66, 0xe6, 0xbd, 0x37, 0x9f, 0xc7, 0x70, 0x74, 0x9d, 0x16, 0x67, 0xab, 0x4a,
	0x28, 0x91, 0xd6, 0x6f, 0xce, 0x24, 0x56, 0xb7, 0x45, 0x86, 0x13, 0x43, 0x90, 0xa1, 0xe7, 0xe9,
	0x05, 0x44, 0x49, 0xb1, 0x44, 0xa9, 0xf8, 0x72, 0x45, 0x62, 0xd8, 0x95, 0x98, 0x89, 0x32, 0x97,
	0x71, 0x70, 0x1c, 0x9c, 0xf6, 0x99, 0x87, 0xe4, 0x03, 0x18, 0x94, 0xbc, 0x14, 0x32, 0xee, 0x19,
	0xde, 0x02, 0xfa, 0x15, 0x90, 0xe9, 0x42, 0x64, 0x37, 0xaf, 0xb0, 0x98, 0x5f, 0x2b, 0x86, 0x6f,
	0x6b, 0x94, 0x8a, 0x7c, 0x0a, 0xe3, 0x54, 0xb3, 0xb3, 0x6b, 0x43, 0xbb, 0xad, 0x46, 0xe9, 0xa6,
	0x92, 0xfe, 0x19, 0xc0, 0x9e, 0x59, 0xc9, 0x50, 0xae, 0x44, 0x29, 0xf1, 0x1e, 0x8b, 0xc8, 0x53,
	0x08, 0x55, 0xb1, 0x44, 0x23, 0x61, 0x74, 0x7e, 0x38, 0xf1, 0x1e, 0x26, 0xad, 0x01, 0x66, 0x0a,
	0xc8, 0x63, 0x88, 0x94, 0x50, 0x7c, 0x31, 0x53, 0x8d, 0x8c, 0xfb, 0xc7, 0xc1, 0x69, 0xc8, 0x86,
	0x86, 0x48, 0x1a, 0x49, 0x9e, 0x03, 0x91, 0xf5, 0x4a, 0xa7, 0x21, 0x45, 0x35, 0xe3, 0x79, 0x5e,
	0xa1, 0x94, 0x71, 0x78, 0x1c, 0x9c, 0x46, 0xec, 0xe1, 0x66, 0xe6, 0xd2, 0x4e, 0xd0, 0x2f, 0xe0,
	0xc1, 0x65, 0x96, 0x89, 0xba, 0x6c, 0xed, 0xc5, 0xb0, 0xeb, 0x57, 0x05, 0x66, 0x95, 0x87, 0xf4,
	0x8f, 0x1e, 0xec, 0xb7, 0xc5, 0xce, 0xd7, 0x9d, 0xd5, 0x26, 0x52, 0x51, 0x66, 0xd6, 0x4f, 0xc8,
	0x2c, 0xd0, 0xf5, 0x29, 0x5f, 0x70, 0xcd, 0x5b, 0xe5, 0x1e, 0xea, 0x84, 0xa4, 0x12, 0x15, 0x9f,
	0xe3, 0xac, 0x12, 0x42, 0x39, 0xc9, 0x23, 0xc7, 0x31, 0x21, 0x14, 0xf9, 0x18, 0x60, 0x55, 0xa7,
	0x8b, 0x22, 0x9b, 0xdd, 0xe0, 0x3a, 0x1e, 0x98, 0x82, 0xc8, 0x32, 0x3f, 0xe0, 0x9a, 0x7c, 0x07,
	0x43, 0xb7, 0x99, 0x8c, 0x77, 0x8e, 0xfb, 0xa7, 0xa3, 0xf3, 0xa7, 0x9b, 0x10, 0xff, 0x21, 0x7c,
	0x32, 0x75, 0x95, 0xdf, 0x97, 0xaa, 0x5a, 0xb3, 0x76, 0xe1, 0xd1, 0x05, 0xec, 0x6d, 0x4d, 0x91,
	0x03, 0xe8, 0xeb, 0xd3, 0xac, 0x3b, 0x3d, 0xd4, 0xce, 0x6e, 0xf9, 0xa2, 0x6e, 0x9d, 0x19, 0xf0,
	0x4d, 0xef, 0xeb, 0x80, 0xfe, 0x1a, 0xc0, 0xe0, 0x52, 0x4a, 0x54, 0xe4, 0x08, 0x86, 0x19, 0x57,
	0x38, 0x17, 0x95, 0x5f, 0xda, 0x62, 0xf2, 0x08, 0x76, 0xe4, 0x7a, 0x99, 0x8a, 0x85, 0xd9, 0x20,
	0x62, 0x0e, 0xe9, 0x6c, 0x4a, 0x54, 0x3f, 0x8b, 0xea, 0xc6, 0x64, 0x13, 0x31, 0x0f, 0x37, 0x27,
	0x86, 0x9d, 0x13, 0xb5, 0xb2, 0x37, 0x88, 0x26, 0x87, 0x90, 0xe9, 0xe1, 0x26, 0xf3, 0x9d, 0x4e,
	0xe6, 0xf4, 0xaf, 0x00, 0x7a, 0x49, 0x43, 0x4e, 0xe0, 0x81, 0xc4, 0x32, 0xc7, 0x4d, 0x57, 0x58,
	0x61, 0x7b, 0x96, 0x75, 0x1d, 0x41, 0x3e, 0x03, 0x47, 0xcc, 0x56, 0x75, 0xaa, 0x9d, 0x5b, 0x91,
	0x63, 0x4b, 0x5e, 0x19, 0x8e, 0x3c, 0x83, 0x83, 0x0a, 0xb3, 0x02, 0x6f, 0x3b, 0xbb, 0x59, 0xcd,
	0xfb, 0x9e, 0xf7, 0xfb, 0x9d, 0xc0, 0x80, 0xeb, 0x48, 0x8c, 0xf6, 0xd1, 0xf9, 0x7e, 0xe7, 0x4a,
	0x34, 0xcd, 0xec, 0xac, 0x36, 0xbf, 0x44, 0x29, 0xf9, 0x1c, 0xdd, 0xc5, 0x7a, 0x48, 0x08, 0x84,
	0xb2, 0x98, 0x97, 0xc6, 0x53, 0xc4, 0xcc, 0x58, 0x73, 0x6a, 0xbd, 0xc2, 0x78, 0xd7, 0x72, 0x7a,
	0x6c, 0x62, 0x55, 0x5c, 0xd5, 0x32, 0x1e, 0xba, 0x58, 0x0d, 0xa2, 0xcf, 0x20, 0x4a, 0x1a, 0xdf,
	0xdd, 0x4f, 0xa0, 0xa7, 0x1a, 0x63, 0x7c, 0x74, 0x3e, 0xee, 0x7c, 0x62, 0x0d, 0xeb, 0xa9, 0x86,
	0xfe, 0x12, 0x00, 0xe8, 0x5a, 0xd7, 0xdc, 0x87, 0x30, 0x50, 0xcd, 0xac, 0xc8, 0x5d, 0x50, 0xa1,
	0x6a, 0x5e, 0xe7, 0x5a, 0xe8, 0x0a, 0xcb, 0xbc, 0x28, 0xe7, 0xee, 0xb1, 0xf0, 0x50, 0x0b, 0x78,
	0x5b, 0x63, 0x8d, 0xb9, 0x89, 0xa2, 0xcf, 0x1c, 0xea, 0x08, 0x0b, 0xbb, 0xc2, 0xee, 0xb6, 0x4c,
	0xbf, 0x85, 0x47, 0x6d, 0xbf, 0xce, 0x0b, 0xa9, 0xb0, 0xf2, 0xfa, 0xdf, 0xb9, 0x9d, 0xe0, 0xdd,
	0xdb, 0xa1, 0x9f, 0xc3, 0x7e, 0xd2, 0xbc, 0x44, 0xc5, 0x8b, 0x85, 0x5f, 0xf7, 0x3e, 0x2b, 0xf4,
	0xb7, 0x00, 0x0e, 0x36, 0x85, 0xff, 0x66, 0xda, 0xc6, 0xd6, 0x7b, 0x7f, 0x6c, 0xe4, 0x05, 0x40,
	0x56, 0x21, 0x57, 0x85, 0x28, 0x5f, 0x2a, 0x63, 0xfe, 0x8e, 0xf7, 0xab, 0x53, 0x46, 0x3e, 0x82,
	0xa1, 0x7d, 0x11, 0x8b, 0xdc, 0xb5, 0xf5, 0xae, 0xc1, 0xaf, 0x73, 0xfa, 0x7b, 0x00, 0xa3, 0xa4,
	0xe2, 0xa5, 0xe4, 0x99, 0x2e, 0x26, 0x14, 0x9c, 0xbf, 0x8e, 0xe7, 0x31, 0xdb, 0xe2, 0xc8, 0x13,
	0x88, 0x74, 0x67, 0x70, 0x55, 0x57, 0xe8, 0x5a, 0x76, 0x43, 0x90, 0x4f, 0x00, 0x2a, 0xcc, 0xb6,
	0x3b, 0xb5, 0xc3, 0xdc, 0xb3, 0x49, 0xe9, 0x05, 0x90, 0x8e, 0x2e, 0x9f, 0xed, 0x89, 0xfe, 0xbc,
	0x5c, 0x4f, 0x7d, 0xd8, 0xb1, 0xdd, 0xa9, 0xec, 0x25, 0x0d, 0x55, 0x70, 0xb8, 0xb5, 0xf8, 0x7f,
	0x69, 0x32, 0x7a, 0x06, 0x87, 0x49, 0x23, 0xa7, 0x6b, 0xf7, 0x39, 0xfe, 0xf7, 0x2b, 0x7f, 0x01,
	0xa3, 0xa4, 0x91, 0xad, 0xbc, 0x2f, 0xa1, 0xaf, 0x7f, 0x33, 0x81, 0x79, 0x4f, 0x8f, 0xba, 0x57,
	0xbf, 0xdd, 0x37, 0x4c, 0x97, 0xd1, 0x1f, 0xe1, 0xb1, 0x3d, 0x4d, 0xc7, 0x75, 0x59, 0xe6, 0xf7,
	0x3d, 0x55, 0xbf, 0x5c, 0xf6, 0x02, 0xec, 0xd5, 0x59, 0x30, 0x7d, 0x0e, 0x0f, 0x33, 0xb1, 0x9c,
	0x5c, 0x63, 0x95, 0x17, 0xb5, 0xb4, 0x87, 0x4f, 0xc7, 0xaf, 0x2c, 0xbc, 0xd2, 0xe8, 0x2a, 0xf8,
	0xa9, 0xfd, 0xd9, 0xa7, 0x3b, 0x66, 0xf4, 0xe2, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0x59, 0x9b,
	0xca, 0x38, 0x1b, 0x08, 0x00, 0x00,
}
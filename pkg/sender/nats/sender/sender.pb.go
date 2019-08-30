// Code generated by protoc-gen-go. DO NOT EDIT.
// source: sender.proto

package sender

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// SendRequest is a request to the service to send token to the specified wallet
type SendRequest struct {
	Service              string   `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	Id                   string   `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	PublicKey            string   `protobuf:"bytes,3,opt,name=publicKey,proto3" json:"publicKey,omitempty"`
	Token                string   `protobuf:"bytes,4,opt,name=token,proto3" json:"token,omitempty"`
	Amount               string   `protobuf:"bytes,5,opt,name=amount,proto3" json:"amount,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SendRequest) Reset()         { *m = SendRequest{} }
func (m *SendRequest) String() string { return proto.CompactTextString(m) }
func (*SendRequest) ProtoMessage()    {}
func (*SendRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_sender_4b0f7177e54d6030, []int{0}
}
func (m *SendRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendRequest.Unmarshal(m, b)
}
func (m *SendRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendRequest.Marshal(b, m, deterministic)
}
func (dst *SendRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendRequest.Merge(dst, src)
}
func (m *SendRequest) XXX_Size() int {
	return xxx_messageInfo_SendRequest.Size(m)
}
func (m *SendRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SendRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SendRequest proto.InternalMessageInfo

func (m *SendRequest) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

func (m *SendRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *SendRequest) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

func (m *SendRequest) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *SendRequest) GetAmount() string {
	if m != nil {
		return m.Amount
	}
	return ""
}

// SendReply is a reply for SendRequest
type SendReply struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error                string   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SendReply) Reset()         { *m = SendReply{} }
func (m *SendReply) String() string { return proto.CompactTextString(m) }
func (*SendReply) ProtoMessage()    {}
func (*SendReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_sender_4b0f7177e54d6030, []int{1}
}
func (m *SendReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendReply.Unmarshal(m, b)
}
func (m *SendReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendReply.Marshal(b, m, deterministic)
}
func (dst *SendReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendReply.Merge(dst, src)
}
func (m *SendReply) XXX_Size() int {
	return xxx_messageInfo_SendReply.Size(m)
}
func (m *SendReply) XXX_DiscardUnknown() {
	xxx_messageInfo_SendReply.DiscardUnknown(m)
}

var xxx_messageInfo_SendReply proto.InternalMessageInfo

func (m *SendReply) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *SendReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

// SentEvent is an event from the service notifying about a wallet sending transaction result
type SentEvent struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error                string   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	Service              string   `protobuf:"bytes,3,opt,name=service,proto3" json:"service,omitempty"`
	Id                   string   `protobuf:"bytes,4,opt,name=id,proto3" json:"id,omitempty"`
	PublicKey            string   `protobuf:"bytes,5,opt,name=publicKey,proto3" json:"publicKey,omitempty"`
	Token                string   `protobuf:"bytes,6,opt,name=token,proto3" json:"token,omitempty"`
	Amount               string   `protobuf:"bytes,7,opt,name=amount,proto3" json:"amount,omitempty"`
	Transaction          string   `protobuf:"bytes,8,opt,name=transaction,proto3" json:"transaction,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SentEvent) Reset()         { *m = SentEvent{} }
func (m *SentEvent) String() string { return proto.CompactTextString(m) }
func (*SentEvent) ProtoMessage()    {}
func (*SentEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_sender_4b0f7177e54d6030, []int{2}
}
func (m *SentEvent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SentEvent.Unmarshal(m, b)
}
func (m *SentEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SentEvent.Marshal(b, m, deterministic)
}
func (dst *SentEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SentEvent.Merge(dst, src)
}
func (m *SentEvent) XXX_Size() int {
	return xxx_messageInfo_SentEvent.Size(m)
}
func (m *SentEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_SentEvent.DiscardUnknown(m)
}

var xxx_messageInfo_SentEvent proto.InternalMessageInfo

func (m *SentEvent) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *SentEvent) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *SentEvent) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

func (m *SentEvent) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *SentEvent) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

func (m *SentEvent) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *SentEvent) GetAmount() string {
	if m != nil {
		return m.Amount
	}
	return ""
}

func (m *SentEvent) GetTransaction() string {
	if m != nil {
		return m.Transaction
	}
	return ""
}

// SentEventReply is a reply for SentEvent
type SentEventReply struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error                string   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SentEventReply) Reset()         { *m = SentEventReply{} }
func (m *SentEventReply) String() string { return proto.CompactTextString(m) }
func (*SentEventReply) ProtoMessage()    {}
func (*SentEventReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_sender_4b0f7177e54d6030, []int{3}
}
func (m *SentEventReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SentEventReply.Unmarshal(m, b)
}
func (m *SentEventReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SentEventReply.Marshal(b, m, deterministic)
}
func (dst *SentEventReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SentEventReply.Merge(dst, src)
}
func (m *SentEventReply) XXX_Size() int {
	return xxx_messageInfo_SentEventReply.Size(m)
}
func (m *SentEventReply) XXX_DiscardUnknown() {
	xxx_messageInfo_SentEventReply.DiscardUnknown(m)
}

var xxx_messageInfo_SentEventReply proto.InternalMessageInfo

func (m *SentEventReply) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *SentEventReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterType((*SendRequest)(nil), "sender.SendRequest")
	proto.RegisterType((*SendReply)(nil), "sender.SendReply")
	proto.RegisterType((*SentEvent)(nil), "sender.SentEvent")
	proto.RegisterType((*SentEventReply)(nil), "sender.SentEventReply")
}

func init() { proto.RegisterFile("sender.proto", fileDescriptor_sender_4b0f7177e54d6030) }

var fileDescriptor_sender_4b0f7177e54d6030 = []byte{
	// 237 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x91, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0x95, 0xb4, 0x71, 0x9b, 0x2b, 0xea, 0x60, 0x21, 0xe4, 0x81, 0xa1, 0xf2, 0xc4, 0xc4,
	0xc2, 0xc8, 0xc2, 0xc2, 0xc4, 0x16, 0x9e, 0x20, 0xb5, 0x6f, 0xb0, 0x28, 0x76, 0xb0, 0xcf, 0x95,
	0x3a, 0xf3, 0x88, 0xbc, 0x10, 0xe2, 0xdc, 0x40, 0x86, 0x30, 0xc0, 0xf8, 0xfd, 0xa7, 0xb3, 0xfe,
	0xcf, 0x07, 0x17, 0x09, 0xbd, 0xc5, 0x78, 0x3b, 0xc4, 0x40, 0x41, 0x8a, 0x42, 0xfa, 0xbd, 0x82,
	0xcd, 0x33, 0x7a, 0xdb, 0xe1, 0x5b, 0xc6, 0x44, 0x52, 0xc1, 0x2a, 0x61, 0x3c, 0x3a, 0x83, 0xaa,
	0xda, 0x55, 0x37, 0x6d, 0x37, 0xa2, 0xdc, 0x42, 0xed, 0xac, 0xaa, 0x39, 0xac, 0x9d, 0x95, 0xd7,
	0xd0, 0x0e, 0x79, 0x7f, 0x70, 0xe6, 0x09, 0x4f, 0x6a, 0xc1, 0xf1, 0x4f, 0x20, 0x2f, 0xa1, 0xa1,
	0xf0, 0x82, 0x5e, 0x2d, 0x79, 0x52, 0x40, 0x5e, 0x81, 0xe8, 0x5f, 0x43, 0xf6, 0xa4, 0x1a, 0x8e,
	0xcf, 0xa4, 0xef, 0xa1, 0x2d, 0x25, 0x86, 0xc3, 0x89, 0x2b, 0x64, 0x63, 0x30, 0x25, 0xae, 0xb0,
	0xee, 0x46, 0xfc, 0x7a, 0x14, 0x63, 0x0c, 0xf1, 0xdc, 0xa2, 0x80, 0xfe, 0xa8, 0x78, 0x9b, 0x1e,
	0x8f, 0xe8, 0xe9, 0xaf, 0xdb, 0x53, 0xe1, 0xc5, 0x9c, 0xf0, 0x72, 0x5e, 0xb8, 0xf9, 0x55, 0x58,
	0xcc, 0x0b, 0xaf, 0xa6, 0xc2, 0x72, 0x07, 0x1b, 0x8a, 0xbd, 0x4f, 0xbd, 0x21, 0x17, 0xbc, 0x5a,
	0xf3, 0x70, 0x1a, 0xe9, 0x07, 0xd8, 0x7e, 0x4b, 0xfd, 0xeb, 0x5f, 0xf6, 0x82, 0x2f, 0x7d, 0xf7,
	0x19, 0x00, 0x00, 0xff, 0xff, 0xee, 0x59, 0x32, 0xd6, 0xf9, 0x01, 0x00, 0x00,
}
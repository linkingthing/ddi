// Code generated by protoc-gen-go. DO NOT EDIT.
// source: dhcp.proto

package pb

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

type DHCPStartReq struct {
	Service              string   `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	ConfigFile           string   `protobuf:"bytes,2,opt,name=configFile,proto3" json:"configFile,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DHCPStartReq) Reset()         { *m = DHCPStartReq{} }
func (m *DHCPStartReq) String() string { return proto.CompactTextString(m) }
func (*DHCPStartReq) ProtoMessage()    {}
func (*DHCPStartReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b4c6fed4d91e328, []int{0}
}

func (m *DHCPStartReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DHCPStartReq.Unmarshal(m, b)
}
func (m *DHCPStartReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DHCPStartReq.Marshal(b, m, deterministic)
}
func (m *DHCPStartReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DHCPStartReq.Merge(m, src)
}
func (m *DHCPStartReq) XXX_Size() int {
	return xxx_messageInfo_DHCPStartReq.Size(m)
}
func (m *DHCPStartReq) XXX_DiscardUnknown() {
	xxx_messageInfo_DHCPStartReq.DiscardUnknown(m)
}

var xxx_messageInfo_DHCPStartReq proto.InternalMessageInfo

func (m *DHCPStartReq) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

func (m *DHCPStartReq) GetConfigFile() string {
	if m != nil {
		return m.ConfigFile
	}
	return ""
}

type DHCPStopReq struct {
	Service              string   `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DHCPStopReq) Reset()         { *m = DHCPStopReq{} }
func (m *DHCPStopReq) String() string { return proto.CompactTextString(m) }
func (*DHCPStopReq) ProtoMessage()    {}
func (*DHCPStopReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b4c6fed4d91e328, []int{1}
}

func (m *DHCPStopReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DHCPStopReq.Unmarshal(m, b)
}
func (m *DHCPStopReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DHCPStopReq.Marshal(b, m, deterministic)
}
func (m *DHCPStopReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DHCPStopReq.Merge(m, src)
}
func (m *DHCPStopReq) XXX_Size() int {
	return xxx_messageInfo_DHCPStopReq.Size(m)
}
func (m *DHCPStopReq) XXX_DiscardUnknown() {
	xxx_messageInfo_DHCPStopReq.DiscardUnknown(m)
}

var xxx_messageInfo_DHCPStopReq proto.InternalMessageInfo

func (m *DHCPStopReq) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

type CreateSubnet4Req struct {
	Service              string   `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	SubnetName           string   `protobuf:"bytes,2,opt,name=subnetName,proto3" json:"subnetName,omitempty"`
	Pools                []string `protobuf:"bytes,3,rep,name=pools,proto3" json:"pools,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateSubnet4Req) Reset()         { *m = CreateSubnet4Req{} }
func (m *CreateSubnet4Req) String() string { return proto.CompactTextString(m) }
func (*CreateSubnet4Req) ProtoMessage()    {}
func (*CreateSubnet4Req) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b4c6fed4d91e328, []int{2}
}

func (m *CreateSubnet4Req) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateSubnet4Req.Unmarshal(m, b)
}
func (m *CreateSubnet4Req) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateSubnet4Req.Marshal(b, m, deterministic)
}
func (m *CreateSubnet4Req) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateSubnet4Req.Merge(m, src)
}
func (m *CreateSubnet4Req) XXX_Size() int {
	return xxx_messageInfo_CreateSubnet4Req.Size(m)
}
func (m *CreateSubnet4Req) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateSubnet4Req.DiscardUnknown(m)
}

var xxx_messageInfo_CreateSubnet4Req proto.InternalMessageInfo

func (m *CreateSubnet4Req) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

func (m *CreateSubnet4Req) GetSubnetName() string {
	if m != nil {
		return m.SubnetName
	}
	return ""
}

func (m *CreateSubnet4Req) GetPools() []string {
	if m != nil {
		return m.Pools
	}
	return nil
}

func init() {
	proto.RegisterType((*DHCPStartReq)(nil), "pb.DHCPStartReq")
	proto.RegisterType((*DHCPStopReq)(nil), "pb.DHCPStopReq")
	proto.RegisterType((*CreateSubnet4Req)(nil), "pb.CreateSubnet4Req")
}

func init() { proto.RegisterFile("dhcp.proto", fileDescriptor_0b4c6fed4d91e328) }

var fileDescriptor_0b4c6fed4d91e328 = []byte{
	// 159 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4a, 0xc9, 0x48, 0x2e,
	0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x52, 0xf2, 0xe0, 0xe2, 0x71, 0xf1,
	0x70, 0x0e, 0x08, 0x2e, 0x49, 0x2c, 0x2a, 0x09, 0x4a, 0x2d, 0x14, 0x92, 0xe0, 0x62, 0x2f, 0x4e,
	0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0x95, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x82, 0x71, 0x85, 0xe4,
	0xb8, 0xb8, 0x92, 0xf3, 0xf3, 0xd2, 0x32, 0xd3, 0xdd, 0x32, 0x73, 0x52, 0x25, 0x98, 0xc0, 0x92,
	0x48, 0x22, 0x4a, 0xea, 0x5c, 0xdc, 0x10, 0x93, 0xf2, 0x0b, 0xf0, 0x1a, 0xa4, 0x94, 0xc4, 0x25,
	0xe0, 0x5c, 0x94, 0x9a, 0x58, 0x92, 0x1a, 0x5c, 0x9a, 0x94, 0x97, 0x5a, 0x62, 0x42, 0xd0, 0xda,
	0x62, 0xb0, 0x3a, 0xbf, 0xc4, 0x5c, 0xb8, 0xb5, 0x08, 0x11, 0x21, 0x11, 0x2e, 0xd6, 0x82, 0xfc,
	0xfc, 0x9c, 0x62, 0x09, 0x66, 0x05, 0x66, 0x0d, 0xce, 0x20, 0x08, 0x27, 0x89, 0x0d, 0xec, 0x43,
	0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xad, 0xa5, 0xe8, 0x7d, 0xef, 0x00, 0x00, 0x00,
}
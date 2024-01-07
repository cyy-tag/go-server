package net

import (
	"encoding/binary"
	"unsafe"

	"google.golang.org/protobuf/proto"
)

const (
	//默认包头长度
	DefaultPacketHeaderSize = int(unsafe.Sizeof(DefaultPacketHeader{}))
	MaxPacketDataSize       = 0x00FFFFFF
)

type PacketHeader interface {
	Len() uint32
	ReadFrom(packetHeaderData []byte)
	WriteTo(packetHeaderData []byte)
}

type DefaultPacketHeader struct {
	// (flags << 24) | len
	// flags [0,255)
	// len [0,16M)
	LenAndFlags uint32
}

func NewDefaultPacketHeader(len uint32, flags uint8) *DefaultPacketHeader {
	return &DefaultPacketHeader{
		LenAndFlags: uint32(flags)<<24 | len,
	}
}

// 包体长度
func (h *DefaultPacketHeader) Len() uint32 {
	return h.LenAndFlags & 0x00FFFFFF
}

// 标记
func (h *DefaultPacketHeader) Flags() uint8 {
	return uint8(h.LenAndFlags >> 24)
}

// 读取长度
func (h *DefaultPacketHeader) ReadFrom(packetHeaderData []byte) {
	h.LenAndFlags = binary.LittleEndian.Uint32(packetHeaderData)
}

// 写入长度
func (h *DefaultPacketHeader) WriteTo(PacketHeader []byte) {
	binary.LittleEndian.PutUint32(PacketHeader, h.LenAndFlags)
}

type Packet interface {
	Command() PacketCommand

	Message() proto.Message

	GetStreamData() []byte
}

type ProtoPacket struct {
	command PacketCommand
	message proto.Message
	data    []byte
}

func NewProtoPacket(command PacketCommand, message proto.Message) *ProtoPacket {
	return &ProtoPacket{
		command: command,
		message: message,
	}
}

func NewProtoPacketWithData(command PacketCommand, data []byte) *ProtoPacket {
	return &ProtoPacket{
		command: command,
		data:    data,
	}
}

func (p *ProtoPacket) Command() PacketCommand {
	return p.command
}

func (p *ProtoPacket) Message() proto.Message {
	return p.message
}

func (p *ProtoPacket) GetStreamData() []byte {
	return p.data
}

package net

import (
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

type ProtoMessageCreator func() proto.Message

type PacketCommand uint16

type ProtoRegister interface {
	Register(command PacketCommand, protoMessage proto.Message)
}

type ProtoCodec struct {
	ProtoPacketBytesEncoder func(protoPacketBytes [][]byte) [][]byte

	ProtoPacketBytesDecoder func(packetData []byte) []byte

	MessageCreatorMap map[PacketCommand]reflect.Type
}

func NewProtoCodec(protoMessageTypeMap map[PacketCommand]reflect.Type) *ProtoCodec {
	codec := &ProtoCodec{
		MessageCreatorMap: protoMessageTypeMap,
	}
	if codec.MessageCreatorMap == nil {
		codec.MessageCreatorMap = make(map[PacketCommand]reflect.Type)
	}
	return codec
}

func (p *ProtoCodec) Register(command PacketCommand, protoMessage proto.Message) {
	if protoMessage == nil {
		p.MessageCreatorMap[command] = nil
	}
	p.MessageCreatorMap[command] = reflect.TypeOf(protoMessage).Elem()
}

func (p *ProtoCodec) Encode(packet Packet) ([]byte, error) {
	encodedData := p.EncodePacket(packet)
	encodedDataLen := 0
	for _, data := range encodedData {
		encodedDataLen += len(data)
	}
	packetHeader := NewDefaultPacketHeader(uint32(encodedDataLen), 0)
	packetHeaderData := make([]byte, DefaultPacketHeaderSize)
	packetHeader.WriteTo(packetHeaderData)
	msgData := make([]byte, DefaultPacketHeaderSize+encodedDataLen)
	copy(msgData, packetHeaderData)
	n := DefaultPacketHeaderSize
	for _, data := range encodedData {
		copy(msgData[n:], data)
		n += len(data)
	}
	return msgData, nil
}

func (p *ProtoCodec) EncodePacket(packet Packet) [][]byte {
	protoMessage := packet.Message()
	commandBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(commandBytes, uint16(packet.Command()))
	var messageBytes []byte
	if protoMessage != nil {
		var err error
		messageBytes, err = proto.Marshal(protoMessage)
		if err != nil {
			fmt.Printf("proto encode err: %v cmd: %v\n", err, packet.Command())
			return nil
		}
	} else {
		messageBytes = packet.GetStreamData()
	}
	return [][]byte{commandBytes, messageBytes}
}

func (p *ProtoCodec) Decode(c gnet.Conn) (Packet, error) {
	buf, _ := c.Peek(DefaultPacketHeaderSize)
	if len(buf) < DefaultPacketHeaderSize {
		return nil, ErrIncompletePacket
	}
	DefaultHeader := DefaultPacketHeader{}
	DefaultHeader.ReadFrom(buf)
	msgLen := DefaultPacketHeaderSize + int(DefaultHeader.Len())
	if c.InboundBuffered() < msgLen {
		return nil, ErrIncompletePacket
	}
	buf, _ = c.Peek(msgLen)
	_, _ = c.Discard(msgLen)
	packet := p.DecodePacket(buf[DefaultPacketHeaderSize:msgLen])
	return packet, nil
}

func (p *ProtoCodec) DecodePacket(packetData []byte) Packet {
	decodedPacketData := packetData
	if len(decodedPacketData) < 2 {
		return nil
	}
	command := binary.LittleEndian.Uint16(decodedPacketData[:2])
	if protoMessageType, ok := p.MessageCreatorMap[PacketCommand(command)]; ok {
		if protoMessageType != nil {
			newProtoMessage := reflect.New(protoMessageType).Interface().(proto.Message)
			err := proto.Unmarshal(decodedPacketData[2:], newProtoMessage)
			if err != nil {
				fmt.Printf("proto decode err:%v cmd:%v\n", err, command)
				return nil
			}
			return &ProtoPacket{
				command: PacketCommand(command),
				message: newProtoMessage,
			}
		} else {
			return &ProtoPacket{
				command: PacketCommand(command),
				data:    decodedPacketData[2:],
			}
		}
	}
	fmt.Printf("unsupport command:%v\n", command)
	return nil
}

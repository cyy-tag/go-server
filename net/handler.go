package net

import (
	"github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

type PacketHandler func(c gnet.Conn, packet *ProtoPacket)

type DefaultConnectionHandler struct {
	//注册消息的处理函数
	PacketHandler map[PacketCommand]PacketHandler
	//解码
	Codec ProtoCodec
	//建立连接消息回调
	onConnectedFunc func(c gnet.Conn)
	//关闭连接消息回调
	onDisconnectedFunc func(c gnet.Conn, err error)
}

func (h *DefaultConnectionHandler) Register(packetCommand PacketCommand, handler PacketHandler, protoMessage proto.Message) {
	h.PacketHandler[packetCommand] = handler
	h.Codec.Register(packetCommand, protoMessage)
}

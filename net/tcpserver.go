package net

import (
	"fmt"

	"github.com/panjf2000/gnet/v2"
)

type TcpServer struct {
	gnet.BuiltinEventEngine
	eng       gnet.Engine
	Handler   DefaultConnectionHandler
	Network   string
	Addr      string
	Multicore bool
}

func (s *TcpServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	fmt.Printf("running server on %s:://%s with multi-core=%t",
		s.Network, s.Addr, s.Multicore)
	s.eng = eng
	return
}

func (s *TcpServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	if s.Handler.onConnectedFunc != nil {
		s.Handler.onConnectedFunc(c)
	}
	out = []byte("sweetness\r\n")
	return
}

func (s *TcpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil {
		fmt.Printf("error occurred on connection=%s, %v\n", c.RemoteAddr().String(), err)
	}
	if s.Handler.onDisconnectedFunc != nil {
		s.Handler.onDisconnectedFunc(c, err)
	}
	return
}

func (s *TcpServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	for {
		packet, err := s.Handler.Codec.Decode(c)
		if err == ErrIncompletePacket {
			return
		}
		if protoPacket, ok := packet.(*ProtoPacket); ok {
			if handler, ok := s.Handler.PacketHandler[packet.Command()]; ok {
				handler(c, protoPacket)
			} else {
				fmt.Printf("can't find register cmd proto %v", packet.Command())
			}
		}
	}
}

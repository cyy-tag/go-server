package net

import (
	"fmt"

	"github.com/panjf2000/gnet/v2"
)

type TcpClient struct {
	gnet.BuiltinEventEngine
	eng          gnet.Engine
	started      int32
	connected    int32
	disconnected int32
	Client       *gnet.Client
	Conn         gnet.Conn
	Handler      DefaultConnectionHandler
	Network      string
	Addr         string
	Nclients     int
}

func (cli *TcpClient) OnBoot(eng gnet.Engine) (action gnet.Action) {
	fmt.Printf("running client on %s:://%s", cli.Network, cli.Addr)
	cli.eng = eng
	return
}

func (cli *TcpClient) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	if cli.Handler.onConnectedFunc != nil {
		cli.Handler.onConnectedFunc(c)
	}
	return
}

func (cli *TcpClient) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil {
		fmt.Printf("error occurred on closed, %v\n", err)
	}
	if cli.Handler.onDisconnectedFunc != nil {
		cli.Handler.onDisconnectedFunc(c, err)
	}
	return
}

func (cli *TcpClient) OnTraffic(c gnet.Conn) (action gnet.Action) {
	packet, err := cli.Handler.Codec.Decode(c)
	if err == ErrIncompletePacket {
		return
	}
	if protoPacket, ok := packet.(*ProtoPacket); ok {
		if handler, ok := cli.Handler.PacketHandler[packet.Command()]; ok {
			handler(c, protoPacket)
		} else {
			fmt.Printf("can't find register cmd proto %v", packet.Command())
		}
	}
	return
}

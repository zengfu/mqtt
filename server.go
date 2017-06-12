package main

import (
	"fmt"
	// "github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/zengfu/backend"
	"github.com/zengfu/packet"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("socket create failed")
	}
	go backend.ChanPool()
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go Session(conn)
	}
}
func Session(socket net.Conn) {
	defer socket.Close()
	var session *packet.ConnectPacket

	sem := make(chan byte, 1)
	stream := packet.NewStream(socket, socket)
	for {
		pkt, err := stream.Decoder.Read()
		if err != nil {
			fmt.Println(err)
			break
		}
		switch pkt.Type() {
		case packet.CONNECT:
			session = pkt.(*packet.ConnectPacket)
			indb := session.SaveDb()

			cnak := packet.NewConnackPacket()
			fmt.Println(session.ID, indb)
			if indb {
				if session.Online {
					cnak.ReturnCode = 2
					cnak.SessionPresent = false
				} else {
					session.UpdateDb()
					cnak.ReturnCode = 0
					if session.CleanSession {
						cnak.SessionPresent = false
					} else {
						cnak.SessionPresent = true
					}
				}
			} else {
				cnak.ReturnCode = 0
				cnak.SessionPresent = false
			}

			//send cnak
			err := stream.Encoder.Write(cnak)
			stream.Encoder.Flush()
			if err != nil {
				fmt.Println(err)
				break
			}
			if cnak.ReturnCode != 0 {
				break
			}
			go backend.AcceptSub(stream, session, sem)
		case packet.SUBSCRIBE:
			sub := pkt.(*packet.SubscribePacket)
			//fmt.Println(sub.PacketID)
			suback := packet.NewSubackPacket()
			for _, sub := range sub.Subscriptions {
				suback.ReturnCodes = append(suback.ReturnCodes, sub.QOS)
			}
			sub.SaveDb(session)
			//go AcceptTopic()
			//have new subscribe
			sem <- 1
			suback.PacketID = sub.PacketID
			err := stream.Encoder.Write(suback)
			stream.Encoder.Flush()
			if err != nil {
				fmt.Println(err)
				break
			}
		case packet.PUBLISH:
			pub := pkt.(*packet.PublishPacket)
			backend.AcceptPub(pub.Message.Topic, pub.Message.Payload)
		case packet.DISCONNECT:
			//DeleteSession(session)
			break
		}
	}
	fmt.Println("socket close")
	session.DeleteDb()
}

package broker

import (
	"fmt"
	"net"
	"time"
)

func ProcessSession(socket net.Conn) {
	defer socket.Close()
	socket.SetReadDeadline(time.Now().Add(1 * time.Second))
	var session *ConnectPacket
	sem := make(chan byte, 1)
	stream := NewStream(socket, socket)
	first := true
	for {
		pkt, err := stream.Decoder.Read()
		if err != nil {
			fmt.Println(err)
			break
		}

		if first {
			session, ok := pkt.(*ConnectPacket)
			if !ok {
				fmt.Println("expected connect")
				break
			}
			err = processConnect(session)
			first = false
		}
		socket.SetReadDeadline(time.Now().Add(time.Duration(float32(session.KeepAlive)*1.5) * time.Second))

		switch _pkt := pkt.(type) {
		case *SubscribePacket:
			err = s.processSubscribe(stream, _pkt)
		case *UnsubscribePacket:
			err = s.processUnsubscribe(stream, _pkt)
		case *PublishPacket:
			err = s.processPublic(stream, _pkt)
		case *PubackPacket:
			err = s.processPuback(stream, _pkt)
		case *PubrecPacket:
			err = s.processPubrec(stream, _pkt)
		case *PubrelPacket:
			err = s.processPubrel(stream, _pkt)
		case *PubcompPacket:
			err = s.processPubcomp(stream, _pkt)
		case *PingreqPacket:
			err = s.processPingreq(stream, _pkt)
		case *DisconnectPacket:
			err = s.processDisconnect(stream, _pkt)
		}
		if err != nil {
			break
		}
	}
	fmt.Println("socket close")

}
func processConnect(session *ConnectPacket) error {
	connack := NewConnackPacket()
	connack.ReturnCode = ConnectionAccepted
	connack.SessionPresent = false
	if session.CleanSession{
		if session.QuerySession(){
			//indb
			session.
		}
	}
	return nil

}
func processSubscribe(stream Stream, sub *SubscribePacket) error {

}
func processUnsubscribe(stream Stream, unsub *UnsubscribePacket) error {

}
func processPublic(stream Stream, public *PublishPacket) error {

}
func processPuback(stream Stream, puback *PubackPacket) error {

}
func processPubrec(stream Stream, pubrec *PubrecPacket) error {

}
func processPubrel(stream Stream, pubrel *PubrelPacket) error {

}
func processPubcomp(stream Stream, pubcomp *PubcompPacket) error {

}
func processPingreq(stream Stream, preq *PingreqPacket) error {

}
func processDisconnect(stream Stream, disconnect *DisconnectPacket) error {

}

// switch _pkt := pkt.(type) {
// 		case PINGREQ:
// 			pingack := NewPingrespPacket()
// 			err := stream.Encoder.Write(pingack)
// 			err = stream.Encoder.Flush()
// 			if err != nil {
// 				fmt.Println(err)
// 				break
// 			}
// 		case CONNECT:
// 			session = pkt.(*ConnectPacket)
// 			indb := session.SaveDb()
// 			cnak := NewConnackPacket()
// 			fmt.Println(session.ID, indb)
// 			if indb {
// 				if session.Online {
// 					cnak.ReturnCode = 2
// 					cnak.SessionPresent = false
// 				} else {
// 					session.UpdateDb()
// 					cnak.ReturnCode = 0
// 					if session.CleanSession {
// 						cnak.SessionPresent = false
// 					} else {
// 						cnak.SessionPresent = true
// 					}
// 				}
// 			} else {
// 				cnak.ReturnCode = 0
// 				cnak.SessionPresent = false
// 			}
// 			defer DeleteDbPool(session.ID, cnak.ReturnCode)
// 			//send cnak
// 			err := stream.Encoder.Write(cnak)
// 			stream.Encoder.Flush()
// 			if err != nil {
// 				fmt.Println(err)
// 				break
// 			}
// 			if cnak.ReturnCode != 0 {
// 				break
// 			}
// 			go AcceptSub(stream, session, sem)
// 		case SUBSCRIBE:
// 			sub := pkt.(*SubscribePacket)
// 			//fmt.Println(sub.PacketID)
// 			suback := NewSubackPacket()
// 			for _, sub := range sub.Subscriptions {
// 				suback.ReturnCodes = append(suback.ReturnCodes, sub.QOS)
// 			}
// 			sub.SaveDb(session)
// 			//go AcceptTopic()
// 			//have new subscribe
// 			sem <- 1
// 			suback.PacketID = sub.PacketID
// 			err := stream.Encoder.Write(suback)
// 			err = stream.Encoder.Flush()
// 			if err != nil {
// 				fmt.Println(err)
// 				break
// 			}
// 		case PUBLISH:
// 			pub := pkt.(*PublishPacket)
// 			AcceptPub(pub.Message.Topic, pub.Message.Payload)
// 		case DISCONNECT:
// 			//DeleteSession(session)
// 			break
// 		}
// 	}
// 	fmt.Println("socket close")

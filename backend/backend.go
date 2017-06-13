package backend

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/zengfu/mqtt/packet"
	"net"
	"time"
)

type SubContent struct {
	id      uint
	topic   string
	channel chan string
}
type PubContent struct {
	topic   string
	payload string
}
type SubTopic map[uint]SubContent

var (
	SubQueue = make(chan SubContent, 100)
	PubQueue = make(chan PubContent, 100)
	SubPool  = make(map[string]SubTopic)
)

func ChanPool() {

	fmt.Println("chan pool start")
	for {
		select {
		case sub := <-SubQueue:
			//have a new subscribe
			//SubPool[sub.topic] = append(SubPool[sub.topic], sub.channel)
			//SubPool[sub.topic] = append(SubPool[sub.topic], sub)
			topic, ok := SubPool[sub.topic]
			if !ok {
				topic = make(SubTopic)
			}

			topic[sub.id] = sub
			SubPool[sub.topic] = topic
		case pub := <-PubQueue:
			subs, ok := SubPool[pub.topic]
			if ok {
				for _, a := range subs {
					//a <- pub.payload
					a.channel <- pub.payload
				}
			}
		}
		fmt.Println("global chan pool", SubPool)
	}
}
func AcceptPub(topic string, payload string) {
	var t PubContent
	t.topic = topic
	t.payload = payload
	PubQueue <- t
}
func AcceptSub(s *packet.Stream, session *packet.ConnectPacket, sem chan byte) error {
	sublist := make(map[string]chan string)
	for {
		<-sem
		db, err := gorm.Open("mysql", "root:71451085Zf*@/test?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			return err
		}
		defer db.Close()
		var client packet.ConnectPacket

		db.Where("client_id=?", session.ClientID).First(&client)
		var Subscriptions []packet.Subscription
		//db.LogMode(true)

		db.Model(&client).Related(&Subscriptions, "Subscriptions")
		//fmt.Println(len())
		//fmt.Println(client.ID, "lllll", len(Subscriptions))
		//db.Model(&client).Related(&Subscriptions)
		for _, read2sub := range Subscriptions {
			//fmt.Println(i, read2sub.Topic)
			_, ok := sublist[read2sub.Topic]
			if !ok {
				payload := make(chan string)
				sublist[read2sub.Topic] = payload
				fmt.Println("go", read2sub.Topic)
				go OneTopic(s, read2sub.Topic, payload)
				var topic SubContent
				topic.id = session.ID
				topic.topic = read2sub.Topic
				topic.channel = payload
				SubQueue <- topic
			}
		}
	}
	return nil
}
func OneTopic(s *packet.Stream, topic string, sem chan string) {
	for {
		payload := <-sem
		//fmt.Println()
		fmt.Println("recev:", topic, payload)
		pub := packet.NewPublishPacket()
		pub.Dup = false
		pub.Message.Topic = topic
		pub.Message.Payload = payload
		err := s.Encoder.Write(pub)
		err = s.Encoder.Flush()
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
func DeleteDbPool(Id uint, returncode packet.ConnackCode) error {
	db, err := gorm.Open("mysql", "root:71451085Zf*@/test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return err
	}
	defer db.Close()
	//db.LogMode(true)
	var client packet.ConnectPacket
	if db.Where("id=?", Id).First(&client).RecordNotFound() {
		return nil
	}
	if returncode == 0 {

		if client.CleanSession {
			var Subscriptions []packet.Subscription
			db.Model(&client).Related(&Subscriptions, "Subscriptions")
			for _, sub := range Subscriptions {
				topic := SubPool[sub.Topic]
				delete(topic, client.ID)
				db.Delete(&sub)
			}
			db.Delete(&client)
		} else {
			client.Online = false
			db.Save(&client)
		}
	}
	return nil
}
func Session(socket net.Conn) {
	defer socket.Close()
	socket.SetReadDeadline(time.Now().Add(1 * time.Second))
	var session *packet.ConnectPacket
	sem := make(chan byte, 1)
	stream := packet.NewStream(socket, socket)

	for {
		pkt, err := stream.Decoder.Read()
		if session != nil {
			socket.SetReadDeadline(time.Now().Add(time.Duration(float32(session.KeepAlive)*1.5) * time.Second))
		}

		if err != nil {
			fmt.Println(err)
			break
		}
		switch pkt.Type() {
		case packet.PINGREQ:
			pingack := packet.NewPingrespPacket()
			err := stream.Encoder.Write(pingack)
			err = stream.Encoder.Flush()
			if err != nil {
				fmt.Println(err)
				break
			}
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
			defer DeleteDbPool(session.ID, cnak.ReturnCode)
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
			go AcceptSub(stream, session, sem)
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
			err = stream.Encoder.Flush()
			if err != nil {
				fmt.Println(err)
				break
			}
		case packet.PUBLISH:
			pub := pkt.(*packet.PublishPacket)
			AcceptPub(pub.Message.Topic, pub.Message.Payload)
		case packet.DISCONNECT:
			//DeleteSession(session)
			break
		}
	}
	fmt.Println("socket close")

}

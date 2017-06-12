package backend

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/zengfu/packet"
)

var (
	SubQueue = make(chan SubContent, 100)
	PubQueue = make(chan PubContent, 100)
)

type SubContent struct {
	SessionId uint
	topic     string
	channel   chan string
}
type PubContent struct {
	topic   string
	payload string
}

func ChanPool() {
	SubPool := make(map[string][]chan string)
	fmt.Println("chan pool start")
	for {
		select {
		case sub := <-SubQueue:
			//have a new subscribe
			SubPool[sub.topic] = append(SubPool[sub.topic], sub.channel)

		case pub := <-PubQueue:
			subs, ok := SubPool[pub.topic]
			if ok {
				for _, a := range subs {
					a <- pub.payload
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
				topic.SessionId = session.ID
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
		fmt.Println("recev:", topic, payload)
	}
}

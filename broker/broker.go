package broker

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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
func AcceptSub(s *Stream, session *ConnectPacket, sem chan byte) error {
	sublist := make(map[string]chan string)
	for {
		<-sem
		db, err := gorm.Open("mysql", "root:71451085Zf*@/test?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			return err
		}
		defer db.Close()
		var client ConnectPacket

		db.Where("client_id=?", session.ClientID).First(&client)
		var Subscriptions []Subscription
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
func OneTopic(s *Stream, topic string, sem chan string) {
	for {
		payload := <-sem
		//fmt.Println()
		fmt.Println("recev:", topic, payload)
		pub := NewPublishPacket()
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
func DeleteDbPool(Id uint, returncode ConnackCode) error {
	db, err := gorm.Open("mysql", "root:71451085Zf*@/test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return err
	}
	defer db.Close()
	//db.LogMode(true)
	var client ConnectPacket
	if db.Where("id=?", Id).First(&client).RecordNotFound() {
		return nil
	}
	if returncode == 0 {

		if client.CleanSession {
			var Subscriptions []Subscription
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

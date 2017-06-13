// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package broker

import (
	"encoding/binary"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"strings"
)

// A Subscription is a single subscription in a SubscribePacket.
type Subscription struct {
	ID        uint `gorm:"primary_key"`
	SessionID uint
	// The topic to subscribe.
	Topic string `gorm:"size:65535;not null`

	// The requested maximum QOS level.
	QOS uint8 `gorm:"not null"`
}

func (s *Subscription) String() string {
	return fmt.Sprintf("%q=>%d", s.Topic, s.QOS)
}

// A SubscribePacket is sent from the client to the server to create one or
// more Subscriptions. The server will forward application messages that match
// these subscriptions using PublishPackets.
type SubscribePacket struct {
	// The subscriptions.
	Subscriptions []Subscription

	// The packet identifier.
	PacketID uint16
}

func (s *SubscribePacket) SaveDb(session *ConnectPacket) error {
	db, err := gorm.Open("mysql", "root:71451085Zf*@/test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return err
	}
	defer db.Close()

	//var Subscriptions []Subscription
	var client ConnectPacket
	db.Where("client_id=?", session.ClientID).First(&client)
	for _, ready2sub := range s.Subscriptions {
		ready2sub.SessionID = session.ID
		err := db.Create(&ready2sub).Error
		if err != nil {
			fmt.Println(err)
		}
	}
	//now the model has subscribed topic

	// 	Subscriptions = append(Subscriptions, read2sub)
	// }
	// db.Save(&Subscriptions)
	// fmt.Println(len(Subscriptions))
	return nil
}

// NewSubscribePacket creates a new SUBSCRIBE packet.
func NewSubscribePacket() *SubscribePacket {
	return &SubscribePacket{}
}

// Type returns the packets type.
func (sp *SubscribePacket) Type() Type {
	return SUBSCRIBE
}

// String returns a string representation of the packet.
func (sp *SubscribePacket) String() string {
	var subscriptions []string

	for _, t := range sp.Subscriptions {
		subscriptions = append(subscriptions, t.String())
	}

	return fmt.Sprintf("<SubscribePacket PacketID=%d Subscriptions=[%s]>",
		sp.PacketID, strings.Join(subscriptions, ", "))
}

// Len returns the byte length of the encoded packet.
func (sp *SubscribePacket) Len() int {
	ml := sp.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (sp *SubscribePacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src[total:], SUBSCRIBE)
	total += hl
	if err != nil {
		return total, err
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", sp.Type(), total+2, len(src))
	}

	// read packet id
	sp.PacketID = binary.BigEndian.Uint16(src[total:])
	total += 2

	// check packet id
	if sp.PacketID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", sp.Type())
	}

	// reset subscriptions
	sp.Subscriptions = sp.Subscriptions[:0]

	// calculate number of subscriptions
	sl := int(rl) - 2

	for sl > 0 {
		// read topic
		t, n, err := readLPString(src[total:], sp.Type())
		total += n
		if err != nil {
			return total, err
		}

		// check buffer length
		if len(src) < total+1 {
			return total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", sp.Type(), total+1, len(src))
		}

		// read qos and add subscription
		sp.Subscriptions = append(sp.Subscriptions, Subscription{Topic: t, QOS: src[total]})
		total++

		// decrement counter
		sl = sl - n - 1
	}

	// check for empty subscription list
	if len(sp.Subscriptions) == 0 {
		return total, fmt.Errorf("[%s] empty subscription list", sp.Type())
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (sp *SubscribePacket) Encode(dst []byte) (int, error) {
	total := 0

	// check packet id
	if sp.PacketID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", sp.Type())
	}

	// encode header
	n, err := headerEncode(dst[total:], 0, sp.len(), sp.Len(), SUBSCRIBE)
	total += n
	if err != nil {
		return total, err
	}

	// write packet it
	binary.BigEndian.PutUint16(dst[total:], sp.PacketID)
	total += 2

	for _, t := range sp.Subscriptions {
		// write topic
		n, err := writeLPString(dst[total:], t.Topic, sp.Type())
		total += n
		if err != nil {
			return total, err
		}

		// write qos
		dst[total] = t.QOS

		total++
	}

	return total, nil
}

// Returns the payload length.
func (sp *SubscribePacket) len() int {
	// packet ID
	total := 2

	for _, t := range sp.Subscriptions {
		total += 2 + len(t.Topic) + 1
	}

	return total
}

package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

type Subscription struct {
	ID        uint `gorm:"primary_key"`
	SessionID uint
	// The topic to subscribe.
	Topic string `gorm:"size:65535;not null`

	// The requested maximum QOS level.
	QOS uint8 `gorm:"not null"`
}
type Message struct {
	ID        uint `gorm:"primary_key"`
	SessionID uint
	// The Topic of the message.
	Topic string `gorm:"size:65535"`

	// The Payload of the message.
	Payload string `gorm:"size:65535"`

	// The QOS indicates the level of assurance for delivery.
	QOS byte

	// If the Retain flag is set to true, the server must store the message,
	// so that it can be delivered to future subscribers whose subscriptions
	// match its topic name.
	Retain bool
}
type ConnectPacket struct {
	ID uint `gorm:"primary_key"`
	// The clients client id.
	ClientID string `gorm:"size:65535;not null;unique_index`

	// The keep alive value.
	KeepAlive uint16 `gorm:"not null"`

	// The authentication username.
	Username string `gorm:"size:65535"`

	// The authentication password.
	Password string `gorm:"size:65535"`

	// The clean session flag.
	CleanSession bool `gorm:"not null"`
	// The Online status
	Online bool `gorm:"not null"`
	// The will message.
	Will *Message `gorm:"ForeignKey:SessionId"`

	// The MQTT version 3 or 4 (defaults to 4 when 0).
	Version byte `gorm:"not null"`
	//
	Subscriptions []Subscription `gorm:"ForeignKey:SessionId"`
}

func main() {
	db, err := gorm.Open("mysql", "root:71451085Zf*@/test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	db.LogMode(true)

	//db.Model(&User{}).AddForeignKey("profile_id", "profiles(id)", "RESTRICT", "RESTRICT")

	db.AutoMigrate(&ConnectPacket{})
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&Subscription{})

	db.Exec(" alter table subscriptions add constraint sub unique (session_id,topic);")
	//db.Where("session_id=?", "test").Delete(Session{})
	//var ts Session
	//fmt.Println(db.Where("session_id=?", "test").Delete(Session{})

	//db.Model(&s).Related(&subs, "Subscribes")

	fmt.Println(time.Now())
}

package main

import (
	"fmt"
	// "github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/zengfu/mqtt/backend"
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
		go backend.Session(conn)
	}
}

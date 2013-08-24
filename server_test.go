package main

import (
	"fmt"
	//"bufio"
	"net"
	"testing"
	"time"
	"labix.org/v2/mgo/bson"
	//"encoding/json"
)

//run the server in another terminal window on port 6666

func TestJoinChannel(t *testing.T) {

	conn, err := net.Dial("tcp", "localhost:6666")
	if err != nil {
		t.Fail()
		// handle error
	} else {
		//message := Message{1,"matt","msg", "hi","en", 12, time.Now()}
		msg, err := bson.Marshal(&Message{1,"matt","msg", "hi","en", 12, time.Now()})
		fmt.Println(msg)
		if err != nil {
			fmt.Println("Failed to marshal", err.Error())
			t.Fail()
		}
		conn.Write(msg)
		//fmt.Fprintf(conn, msg)
	}
}

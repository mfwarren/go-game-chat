package main

import (
	"bufio"
	"bytes"
	"fmt"
	"labix.org/v2/mgo/bson"
	"net"
	"os"
	"time"
)

type Message struct {
	UserID        int
	UserName      string
	Command       string
	Content       string
	Language      string
	MessageNumber int
	Date          time.Time
	ChatRoomName  string
}

var running bool // global variable if client is running
var name string

// clientsender(): read from stdin and send it via network
func clientsender(cn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("you> ")
		input, _ := reader.ReadBytes('\n')
		if bytes.Equal(input, []byte("/quit\n")) {
			msg, _ := bson.Marshal(&Message{1, name, "leave", "", "en", 12, time.Now(), "global"})
			cn.Write(msg)
			running = false
			break
		}
		fmt.Println("clientsender(): send: ", string(input[0:len(input)-1]))
		msg, _ := bson.Marshal(&Message{1, name, "msg", string(input[0 : len(input)-1]), "en", 12, time.Now(), "global"})
		cn.Write(msg)
	}
}

// clientreceiver(): wait for input from network and print it out
func clientreceiver(cn net.Conn) {
	//buf := make([]byte, 4096)
	for running {
		//cn.Read(buf)
		buf := make([]byte, 4096)
		_, err := cn.Read(buf)
		if err != nil {
			fmt.Println("error reading from connection")
			return
		}

		message := &Message{}
		err = bson.Unmarshal(buf, message)
		if err != nil {
			fmt.Println("error unmarshaling")
		} else {
			fmt.Println(message)
		}
		fmt.Print("you> ")
	}
}

func main() {
	running = true

	conn, err := net.Dial("tcp", "localhost:6666")
	if err == nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Your Name> ")
		input, _ := reader.ReadBytes('\n')
		name = string(input)

		// join global chat room
		msg, err := bson.Marshal(&Message{1, name, "join", "", "en", 12, time.Now(), "global"})

		if err != nil {
			fmt.Println("Failed to marshal", err.Error())
		}
		conn.Write(msg)

		// start receiver and sender
		go clientreceiver(conn)
		go clientsender(conn)

		// wait for quiting (/quit). run until running is true
		for running {
			time.Sleep(1 * 1e9)
		}

		//TODO  handle error
	} else {
		fmt.Println("Failed to connect")
		//message := Message{1,"matt","msg", "hi","en", 12, time.Now()}
		//fmt.Fprintf(conn, msg)
	}
	defer conn.Close()

}

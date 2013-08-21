/*
Team Chat is a low level high performance chat room service intended to be baked
into mobile games to support team live communication.

a listening goroutine will sit on the incoming messages. it will pass off the
message for processing, which will determine if there are any commands in it or
if it is a message it will relay it back to everyone else in the channel

	possible commands are:
	join
	leave
	request msgs #x - #y  //allow for negatives to index from the end
	ack			//clients need to ack every minute to stay connected to the channels

version History:
v.1  - relay live message back to all listeners on a channel, no history is saved.
v.2 - history is saved. users are authenticated and can be banned
v.3 - additional commands to support regions and languages

Written By Matt Warren 2013
*/

package main

import (
	//"fmt"
	"net"
	"os"
	//"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	//"encoding/json"
	//"strconv"
	//"strings"
	"time"
)

const (
	PORT string = ":6666" //default port
	RECV_BUF_LEN = 1024
)

type Server struct {
	connection *net.Conn  //listening on
	messages   chan string		//incoming messages/commands
}

type Channel struct {
	messages   chan string		//outgoing messages to channel
	clients    map[int]Client	//clients listening to Channel
}

type Client struct {
	userID   int
	userName string
	userAddr *net.Addr
}

type Message struct {
	userID   int
	userName string
	command  string
	content  string
	language string
	messageNumber int
	date     time.Time
}

func handleMessage(conn net.Conn) {

	//read data coming in - should be bson formatted Message
	buf := make([]byte, RECV_BUF_LEN)
	n, err := conn.Read(buf)
	if err != nil {
		println("Error reading:", err.Error())
		return
	}
	println("received ", n, " bytes of data =", string(buf))
	var message Message
	err = bson.Unmarshal(buf, &message)
	if err != nil {
		println("ERROR:", err.Error())
	}

	//analyse message



	//send reply
	_, err = conn.Write(buf)
	if err != nil {
		println("Error send reply:", err.Error())
	}else {
		println("Reply sent")
	}
}

func main() {

	var s Server
	s.messages = make(chan string, 30)

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		println("Error listening", err.Error())
		os.Exit(1)
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			println("error accept:", err.Error())
			continue
		}
		go handleMessage(conn)
	}

}

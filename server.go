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
	ping			//clients need to ping every minute to stay connected to the channels
	flag msg #		//flag msgs for inappropriate language or spam

version History:
v.1  - relay live message back to all listeners on a channel, no history is saved.
v.2 - history is saved. users are authenticated and can be banned
v.3 - additional commands to support regions and languages

Written By Matt Warren 2013
*/

package main

import (
	"fmt"
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
	PORT         string = ":6666" //default port
	RECV_BUF_LEN        = 1024
)

type Channel struct {
	messages chan Message   //outgoing messages to channel
	clients  map[int]Client //clients listening to Channel
}

type Client struct {
	userID   int
	userName string
	userAddr *net.Addr
}

type Message struct {
	UserID        int
	UserName      string
	Command       string
	Content       string
	Language      string
	MessageNumber int
	Date          time.Time
}

func (m *Message) String() string {
	return fmt.Sprintf("%s %s %s", m.Command, m.UserName, m.Content)
}

func parseMessage(conn net.Conn) (message *Message, err error) {
	buf := make([]byte, RECV_BUF_LEN)
	_, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}
	//println("received ", n, " bytes of data =", string(buf))

	err = bson.Unmarshal(buf, &message)
	if err != nil {
		return nil, err
	}
	fmt.Println(message)
	return message, nil
}

func handleMessage(conn net.Conn) {

	//read data coming in - should be bson formatted Message
	//var message Message
	message, err := parseMessage(conn)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}

	//analyse message
	switch message.Command {
	case "msg":
		fmt.Println("Got a Msg")
	case "join":
		fmt.Println("Join Channel")
	case "leave":
		fmt.Println("Leave Channel")
	case "ping":
		fmt.Println("Ping")
	case "request":
		fmt.Println("Request Messages")
	case "flag":
		fmt.Println("Flag Message")
	default:
		fmt.Println("unknown command: ", message)
	}
}

func main() {

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		println("Error listening", err.Error())
		os.Exit(1)
	}

	defer listener.Close()

	//channel := make(Channel)

	for {
		conn, err := listener.Accept()
		if err != nil {
			println("error accept:", err.Error())
		} else {
			go handleMessage(conn) //pull in from TCP and push on to server Messages channel
		}
	}

}

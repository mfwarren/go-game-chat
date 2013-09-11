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
v.1 - connect, join channel, msg and leave should be implemented.
v.2 - history is saved. users are authenticated and can be banned
v.3 - additional commands to support regions and languages

Written By Matt Warren 2013

//TODO
reader/writer with fan-in concurrency pattern
each connection will join one chat room,
a reader will read from the connection and forward it onto a read channel for the room
a process will read from the read channel, process the message, log it, and write to write channel
a goroutine per channel will read from the write channel, loop through all the clients and write to their connections

a new chatroom will spin up a reader and writer go-routine
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

type Server struct {
	ChatRooms map[string]*ChatRoom
}

type ChatRoom struct {
	WriteMessages chan *Message      //outgoing messages to channel
	Clients       map[string]*Client //clients listening to Channel
	Name          string
	ReadChannel   chan *Message
}

type Client struct {
	UserID   int
	UserName string
	UserConn net.Conn
	LastPing time.Time
}

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

func AppendClient(slice []*Client, data ...*Client) []*Client {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]*Client, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func (m *Message) String() string {
	return fmt.Sprintf("%s %s %s", m.Command, m.UserName, m.Content)
}
func (c *Client) String() string {
	return fmt.Sprintf("%s %s %s", c.UserName, c.UserConn, c.LastPing)
}
func (c *ChatRoom) RegisterClient(client *Client) {
	c.Clients[client.UserName] = client
	//TODO:
	// spin off reader goroutine
	go clientReader(client.UserConn, c.ReadChannel)
	// run writer goroutine
}

func chatroomWriter(chatroom *ChatRoom) {
	// go through all connections and write back to them
	// TODO add quit channel for when chatroom closes
	for {
		msg := <-chatroom.WriteMessages
		marshalled_msg, err := bson.Marshal(msg)

		if err != nil {
			fmt.Println("Failed to marshal", err.Error())
		}

		for username, client := range chatroom.Clients {
			//client := chatroom.Clients[i]
			fmt.Println("sending msg to ", username)
			client.UserConn.Write(marshalled_msg)
		}
	}
}

func clientReader(conn net.Conn, readChan chan *Message) {
	// parse messages from client push into reader channel for processing
	// TODO add quit channel to exit this loop
	for {
		message, err := parseMessage(conn)
		if err != nil {
			fmt.Println("Error parsing Message: ", err.Error())
			continue
		}
		readChan <- message
	}
}

func chatRoomReader(readChan chan *Message, writeChan chan *Message) {
	//TODO do something based on what comes in.
	// send to writer channel which will push strings to clients
	// TODO add quit channel to close a room when nobody's in it.
	var msg *Message
	for {
		msg = <-readChan

		switch msg.Command {
		case "msg":
			fmt.Println("Got a Msg")
			writeChan <- msg
		case "join":
			fmt.Println("Join Chatroom")
		case "leave":
			fmt.Println("Leave Chatroom")
		case "ping":
			fmt.Println("Ping")
		case "request":
			fmt.Println("Request Messages")
		case "flag":
			fmt.Println("Flag Message")
		default:
			fmt.Println("unknown command: ", msg)
		}

	}
}

func parseMessage(conn net.Conn) (message *Message, err error) {
	//var buf [4048]byte
	buf := make([]byte, RECV_BUF_LEN)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println("error reading from connection")
		return nil, err
	}

	message = &Message{}
	err = bson.Unmarshal(buf, message)
	if err != nil {
		fmt.Println("error unmarshaling")
		return nil, err
	}
	fmt.Println(message)
	return message, nil
}

func joinChat(server *Server, conn net.Conn) {

	//read data coming in - should be bson formatted Message
	//var message Message
	message, err := parseMessage(conn)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}

	fmt.Println("JOIN CHAT")

	//analyse message
	if message.Command == "join" {
		fmt.Println("Join Chatroom")
		//create a Client struct for the connection
		client := Client{UserID: message.UserID, UserName: message.UserName, UserConn: conn, LastPing: time.Now()}
		fmt.Println(client)

		//create the channel if  doesn't exist
		//start channel goroutine
		var val *ChatRoom
		val, ok := server.ChatRooms[message.ChatRoomName]
		if !ok {
			fmt.Println("CREATING NEW CHATROOM")
			read_channel := make(chan *Message)
			write_channel := make(chan *Message)
			clients_list := make(map[string]*Client)
			val = &ChatRoom{write_channel, clients_list, message.ChatRoomName, read_channel}
			server.ChatRooms[message.ChatRoomName] = val
			go chatRoomReader(read_channel, write_channel)
			go chatroomWriter(val)
		}
		val.RegisterClient(&client)
	} else {
		fmt.Println("unknown command: ", message)
	}
}

func main() {

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		println("Error listening", err.Error())
		os.Exit(1)
	}

	rooms := make(map[string]*ChatRoom)
	server := Server{ChatRooms: rooms}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			println("error accept:", err.Error())
		} else {
			// new user connecting.
			go joinChat(&server, conn) //pull in from TCP and push on to server Messages channel
		}
	}

}

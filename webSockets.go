package main

import (
	"fmt"
	types "my-social-network/types"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	user           *types.User
	conn           *websocket.Conn
	messageChannel chan []byte
}

var clients = MakeMap()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func addClient(user types.User, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := Client{
		user:           &user,
		conn:           ws,
		messageChannel: make(chan []byte),
	}

	clients.Set(user.Id, &client)

	go writeMessage(user.Id)
	go readMessages(user.Id)
}

func removeClient(id int) {
	clients.Delete(id)

}

func readMessages(id int) {
	client := clients.Get(id)
	if client == nil {
		return
	}

	defer func() {
		removeClient(id)
	}()
	for {
		_, message, err := client.conn.ReadMessage()
		fmt.Println("Incoming message: ", string(message))
		if err != nil {
			// Error:  websocket: close 1001 (going away)
			fmt.Println(err, " Connection: ", id)
			return
		}
	}
}

func writeMessage(id int) {
	client := clients.Get(id)
	if client == nil {
		return
	}

	defer func() {
		removeClient(id)
	}()
	for {
		select {
		case message, ok := <-client.messageChannel:
			if ok {
				if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					fmt.Println(err)
					return
				}
			} else {
				if err := client.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func notifyClient(id int, message []byte) {
	client := clients.Get(id)
	if client != nil {
		client.messageChannel <- message
	}
}

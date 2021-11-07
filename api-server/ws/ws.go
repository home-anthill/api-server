package ws

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

var ids []string

// use go send() to send a message
func Send() {
	hubInstance := GetInstance()
	fmt.Println("ids: ", ids)
	if len(ids) != 0 {
		res0 := hubInstance.sendToClient(ids[0], Message{Type: 1, Body: "only to user 1"})
		if res0 {
			fmt.Println("Message 0 sent")
		}
	}
	if len(ids) > 1 {
		res1 := hubInstance.sendToClient(ids[1], Message{Type: 2, Body: "only to user 2"})
		if res1 {
			fmt.Println("Message 1 sent")
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	hub := GetInstance()
	// create a new client with data received from HTTP request
	// in this case to do a basic example I generate a random UUID
	// to identify this client
	newId := uuid.NewString()
	ids = append(ids, newId)
	client := &Client{
		ID: newId,
		Hub: hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}
	// register the new client to the hub
	client.Hub.Register <- client
}
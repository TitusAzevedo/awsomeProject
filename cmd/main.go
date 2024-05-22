package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

type ClientMessage struct {
	Message Message
	*websocket.Conn
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan ClientMessage)

func main() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Welcome to the Chat Room!")
	if err != nil {
		return
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		var msg Message

		err := conn.ReadJSON(&msg)

		if err != nil {
			fmt.Println(err)
			delete(clients, conn)
			return
		}

		broadcast <- ClientMessage{
			Message: msg,
			Conn:    conn,
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client := range clients {
			if client != msg.Conn {
				err := client.WriteJSON(msg.Message)
				if err != nil {
					fmt.Println(err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

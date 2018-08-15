package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

var upgrader = websocket.Upgrader{}

//Message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	//upgrade GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	//close the connection when funcion returns
	defer ws.Close()
	//Register new Client
	clients[ws] = true

	for {
		var msg Message

		//read new msg as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		//sending msg on broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		//get msg from broadcast
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	//Default server path
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	//Route for connecion to websocket
	http.HandleFunc("/ws", handleConnections)

	//go routine for handling messages.
	go handleMessages()

	log.Println("http server is started on :8088")
	//Server Listening on port 8088
	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		log.Fatal("Listen and Serve: ", err)
	}
}

package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}

func WsEndpoint(w http.ResponseWriter, r *http.Request, ws *websocket.Conn) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	// upgrade this connection to a WebSocket
	// connection
	ws, _ = upgrader.Upgrade(w, r, nil)
	// if err != nil {
	//  log.Println(err)
	// }
	log.Println("Client Connected")
	err1 := ws.WriteMessage(1, []byte("{}"))
	if err1 != nil {
		log.Println(err1)
	}
	reader(ws)
}

func WriteMessage(text []byte, ws *websocket.Conn) error {
	return ws.WriteMessage(websocket.TextMessage, text)
}

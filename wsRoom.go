package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Broadcast struct {
	MessageCh chan string
	clients   map[*websocket.Conn]bool
	roomMu    sync.Mutex
	stop      chan struct{}
}

func newBroadcast() *Broadcast {

	return &Broadcast{
		MessageCh: make(chan string),
		clients:   make(map[*websocket.Conn]bool),
		stop:      make(chan struct{}),
	}
}

func (b *Broadcast) run() {

	for {
		select {

		case msg := <-b.MessageCh:
			b.roomMu.Lock()

			for client := range b.clients {
				err := client.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Println("Error: ", err)
					client.Close() //close the client underlying connection (no point taking to a client with a broken connection)
					delete(b.clients, client)
				}
			}

			b.roomMu.Unlock()
		case <-b.stop:
			log.Println("Stopping ...")
			return
		}
	}

}

func (b *Broadcast) WShandle(c echo.Context) error {

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), c.Response().Header())
	if err != nil {
		log.Println("Failed to upgrade to websocket: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{

			"message": "something went wrong1",
		})
	}
	defer ws.Close()

	b.roomMu.Lock()
	b.clients[ws] = true
	b.roomMu.Unlock()

	for {

		_, msg, err := ws.ReadMessage()
		if err != nil {
			b.roomMu.Lock()
			delete(b.clients, ws)
			b.roomMu.Unlock()
			log.Println("Error: client disconnected! : ", err)
			break
		}

		b.MessageCh <- string(msg)
	}

	return c.JSON(http.StatusOK, "")
}

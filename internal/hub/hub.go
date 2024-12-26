package hub

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	conn *websocket.Conn
	msg  []byte
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	Send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("[CLIENT %s] Received message: %s\n", c.conn.RemoteAddr(), message)
		c.hub.Broadcast <- Message{c.conn, message}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			log.Printf("[CLIENT %s] Sending message: %s\n", c.conn.RemoteAddr(), message)
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func Serve(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading connection: %v", err)
		return
	}
	client := &Client{hub: hub, conn: conn, Send: make(chan []byte, 256)}
	client.hub.Register <- client

	log.Printf("Serving client: %s\n", conn.RemoteAddr())

	go client.writePump()
	go client.readPump()
}

type Hub struct {
	clients map[*Client]bool

	Broadcast chan Message

	Register chan *Client

	Unregister chan *Client
}

func New() *Hub {
	return &Hub{
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			log.Printf("Registering client: %s\n", client.conn.RemoteAddr())
			h.clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				log.Printf("Unregistering client: %s\n", client.conn.RemoteAddr())
				delete(h.clients, client)
				close(client.Send)

				if len(h.clients) == 0 {
					return
				}
			}
		case message := <-h.Broadcast:
			log.Printf("Broadcasting message: %s\n", message.msg)
			for client := range h.clients {
				if client.conn != message.conn {
					client.Send <- message.msg
				}
			}
		}
	}
}

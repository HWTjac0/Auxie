package handlers

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type WSClient struct {
	Conn     *websocket.Conn
	UserID   int
	Username string
	Role     string
	Send     chan interface{}
}

func (c *WSClient) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for msg := range c.Send {
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			break
		}
	}
}

type Subscription struct {
	RoomID string
	Client *WSClient
}

type BroadcastMessage struct {
	RoomID  string
	Payload interface{}
}

type RoomHub struct {
	// roomID -> clients
	rooms       map[string]map[*WSClient]bool
	broadcast   chan *BroadcastMessage
	register    chan *Subscription
	unregister  chan *Subscription
	onUserJoin  func(roomID string, client *WSClient)
	onUserLeave func(roomID string, client *WSClient)
}

func NewRoomHub() *RoomHub {
	h := &RoomHub{
		rooms:      make(map[string]map[*WSClient]bool),
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Subscription),
		unregister: make(chan *Subscription),
	}
	go h.run()
	return h
}

func (h *RoomHub) run() {
	for {
		select {
		case sub := <-h.register:
			clients := h.rooms[sub.RoomID]
			if clients == nil {
				clients = make(map[*WSClient]bool)
				h.rooms[sub.RoomID] = clients
			}

			userConnectionCount := 0
			for client := range clients {
				if client.UserID == sub.Client.UserID {
					userConnectionCount++
				}
			}

			clients[sub.Client] = true

			if userConnectionCount == 0 && h.onUserJoin != nil {
				go h.onUserJoin(sub.RoomID, sub.Client)
			}

		case sub := <-h.unregister:
			clients := h.rooms[sub.RoomID]
			if clients != nil {
				if _, ok := clients[sub.Client]; ok {
					delete(clients, sub.Client)
					close(sub.Client.Send)

					userConnectionCount := 0
					for client := range clients {
						if client.UserID == sub.Client.UserID {
							userConnectionCount++
						}
					}

					if userConnectionCount == 0 && h.onUserLeave != nil {
						go h.onUserLeave(sub.RoomID, sub.Client)
					}

					if len(clients) == 0 {
						delete(h.rooms, sub.RoomID)
					}
				}
			}

		case msg := <-h.broadcast:
			clients := h.rooms[msg.RoomID]
			for client := range clients {
				select {
				case client.Send <- msg.Payload:
				default:
					// Slow client: drop it to prevent blocking the hub
					delete(clients, client)
					close(client.Send)
					client.Conn.Close()
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

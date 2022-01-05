package hubconn

import (
	"github.com/gorilla/websocket"
)

type Connection struct {
	ws   *websocket.Conn
	send chan []byte
}

type Messgae struct {
	data []byte
	room string
}

type Hub struct {
	rooms      map[string]map[*Connection]bool
	broadcast  chan Messgae
	register   chan Subscription
	unregister chan Subscription
}

var H = Hub{
	rooms:      make(map[string]map[*Connection]bool),
	broadcast:  make(chan Messgae),
	register:   make(chan Subscription),
	unregister: make(chan Subscription),
}

func NewHub() *Hub {
	return &H
}

func (x *Hub) Run() {
	for {
		select {
		case s := <-x.register:
			connnections := x.rooms[s.room]
			if connnections == nil {
				connnections = make(map[*Connection]bool)
				x.rooms[s.room] = connnections
			}
			x.rooms[s.room][s.con] = true
		case s := <-x.unregister:
			connections := x.rooms[s.room]
			if connections != nil {
				// underscore has index, ok has boolean value
				if _, ok := connections[s.con]; ok {
					delete(connections, s.con)
					close(s.con.send) // closing the channel
					if len(connections) == 0 {
						delete(x.rooms, s.room)
					}
				}
			}

		// here type message struct is used
		case m := <-x.broadcast:
			connections := x.rooms[m.room]
			for c := range connections {
				select {
				case c.send <- m.data:
				default:
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(x.rooms, m.room)
					}
				}
			}
		}
	}
}

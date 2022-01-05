package hubconn

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second // time allowed to write a message to the peer, time given to backend, we use select case here for timer

	pongWait = 60 * time.Second // time wlloed to read the next pong message from the peer

	pingPeriod = (pongWait * 9) / 10 // send pings to peer with their period. Must be less than pong wait

	maxMessageSize = 512
)

type Subscription struct {
	room string
	con  *Connection
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Subscription) Readpump() {
	c := s.con
	defer func() {
		H.unregister <- *s
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))

	// have to comment this line of code and run to see what it really does
	c.ws.SetPongHandler(
		func(appData string) error {
			c.ws.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		m := Messgae{
			data: msg,
			room: s.room,
		}
		H.broadcast <- m
	}
}

func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (s *Subscription) WritePump() {
	c := s.con
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{}) // empty message and a byte string indicating closed message flag is sent
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func ServeWs(w http.ResponseWriter, r *http.Request, roomId string) {
	fmt.Println(roomId)

	wsx, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	c := &Connection{
		ws:   wsx,
		send: make(chan []byte),
	}

	s := Subscription{
		room: roomId,
		con:  c,
	}
	H.register <- s
	go s.Readpump()
	go s.WritePump()
}

package bbs

import (
	"encoding/json"
	"fmt"

	"code.google.com/p/go.net/websocket"
)

const sendQueueSize = 10

type client struct {
	srv    *Server
	socket *websocket.Conn
	sesh   *Session

	sendq chan interface{}
}

func newClient(srv *Server, socket *websocket.Conn) *client {
	return &client{
		srv:    srv,
		socket: socket,
		sendq:  make(chan interface{}, sendQueueSize),
		sesh:   &Session{BBS: srv.NewBBS()},
	}
}

func (c *client) Send(msg interface{}) {
	c.sendq <- msg
}

func (c *client) writer() {
	for {
		select {
		case msg, ok := <-c.sendq:
			if !ok {
				// our work here is done
				return
			}
			err := websocket.JSON.Send(c.socket, msg)
			if err != nil {
				// disconnect etc
				return
			}
		}
	}
}

func (c *client) run() {
	defer c.cleanup()
	if r, ok := c.sesh.BBS.(Realtime); ok {
		r.Connect(c)
	}
	for {
		var data []byte
		err := websocket.Message.Receive(c.socket, &data)
		if err != nil {
			break
		}

		incoming := BBSCommand{}
		err = json.Unmarshal(data, &incoming)
		if err != nil {
			fmt.Println("JSON Parsing Error!! " + string(data))
			continue
		}
		result := c.srv.do(incoming, data, c.sesh)
		/*
			switch result := result.(type) {
			case WelcomeMessage:

			}
		*/
		c.Send(result)
	}
}

func (c *client) cleanup() {
	// post-disconnect cleanup
	c.socket.Close()
	close(c.sendq)
	if c.sesh != nil {
		if r, ok := c.sesh.BBS.(Realtime); ok {
			r.Bye()
		}
	}
}

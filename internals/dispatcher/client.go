package dispatcher

import (
	"goftp/logging"
	"net"
)

type Client struct {
	terminate chan struct{}
	log       *logging.Client
}

func NewClient(log *logging.Client) *Client {
	return &Client{terminate: make(chan struct{}, 1), log: log}
}

func (c *Client) Start() {

	c.log.Infof("Dispatcher starting")

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}

	for {
		select {
		case <-c.terminate:
			// exit loop
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				// handle error
			}
			go c.handleConnection(conn)

		}
	}

}

func (c *Client) Stop() {
	c.log.Infof("Dispatcher stopping")
	c.terminate <- struct{}{}
	close(c.terminate)
}

func (c *Client) handleConnection(conn net.Conn) {
	c.log.Infof("Connection recv")
	conn.Write([]byte("response\n"))
	conn.Close()
	c.log.Infof("Connection handled")
}

package dispatcher

import (
	"goftp/logging"
	"net"
)

type Client struct {
	terminate chan struct{}
	work      chan net.Conn
	log       *logging.Client
}

func NewClient(log *logging.Client) *Client {
	return &Client{terminate: make(chan struct{}, 1), log: log, work: make(chan net.Conn, 10)}
}

func (c *Client) Start() {

	c.log.Infof("Dispatcher starting")

	go func() {

		ln, err := net.Listen("tcp", ":8080")
		if err != nil {
			// handle error
		}

		for {
			conn, _ := ln.Accept()
			c.work <- conn
		}
	}()

	for {
		select {
		case <-c.terminate:
			// exit loop
			c.log.Infof("Terminate Recv")
			return
		case work := <-c.work:
			go c.handleConnection(work)
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

package dispatcher

import (
	"fmt"
	"goftp/logging"
	"net"
)

type Client struct {
	log    *logging.Client
	server net.Listener
}

func NewClient(log *logging.Client) *Client {
	return &Client{log: log}
}

// Start kicks off the reactor loop for client control connections initiated by some
// ftp client, essentially a tcp server that is expecting the telnet protocol
func (c *Client) Start() {
	var (
		err  error
		conn net.Conn
	)

	c.log.Infof("Dispatcher starting")

	c.server, err = net.Listen("tcp", ":23")
	if err != nil {
		panic(err.Error())
	}

	for {
		c.log.Infof("waiting for conn")
		conn, err = c.server.Accept()
		if err != nil {
			fmt.Println(err.Error())
			c.log.Infof("Stopping server")
			return
		}
		go c.handleConnection(conn)
	}

}

func (c *Client) Stop() {
	c.log.Infof("Dispatcher stopping")
	c.server.Close()

}

func (c *Client) handleConnection(conn net.Conn) {
	c.log.Infof("Connection recv")
	var buffer []byte
	_, _ = conn.Read(buffer)
	fmt.Printf(string(buffer))
	conn.Write([]byte("response\n"))
	conn.Close()
	c.log.Infof("Connection handled")
}

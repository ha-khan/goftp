package dispatcher

import (
	"fmt"
	"goftp/components/logger"
	"goftp/components/worker"
	"net"
)

type Client struct {
	logger logger.Client
	server net.Listener
	port   string
}

func NewClient(log logger.Client) *Client {
	return &Client{
		logger: log,
	}
}

// Start kicks off the reactor loop for each control connections initiated by some ftp client
func (c *Client) Start() {
	c.logger.Infof("Dispatcher starting")

	var err error
	c.server, err = net.Listen("tcp", ":2023")
	if err != nil {
		panic(err.Error())
	}

	for {
		c.logger.Infof("waiting for conn")

		conn, err := c.server.Accept()
		if err != nil {
			// write back to connection that error with server

			// TODO: need to rethink this scenario, main thread would still be blocking, should
			//       probably throw panic to kill process
			fmt.Println(err.Error())
			c.logger.Infof("Stopping server")

			return
		}

		go worker.NewWorker(c.logger).Start(conn)
	}
}

func (c *Client) Stop() {
	c.logger.Infof("Dispatcher stopping")
	// todo, need to keep track of all outstanding workers
	// which themselves have a connection that they are processing
	// can close gracefully or keep them alive until the client closes them
	// regardless the dispatcher needs to stop accepting new connections at a
	// minimum
	c.server.Close()
}

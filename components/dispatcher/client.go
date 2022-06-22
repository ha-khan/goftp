package dispatcher

import (
	"fmt"
	"goftp/components/logger"
	"net"
)

type Client struct {
	logger logger.Client
	server net.Listener
	port   string
	worker func(logger logger.Client, conn net.Conn)
}

func NewClient(log logger.Client) *Client {
	return &Client{
		logger: log,
		worker: worker,
	}
}

// Start kicks off the reactor loop for client control connections initiated by some
// ftp client, essentially a tcp server that is expecting the telnet protocol
func (c *Client) Start() {
	var (
		err  error
		conn net.Conn
	)

	c.logger.Infof("Dispatcher starting")

	c.server, err = net.Listen("tcp", ":2023")
	if err != nil {
		panic(err.Error())
	}

	for {
		c.logger.Infof("waiting for conn")
		conn, err = c.server.Accept()
		if err != nil {
			// TODO: need to rethink this scenario, main thread would still be blocking, should
			//       probably throw panic to kill process
			fmt.Println(err.Error())
			c.logger.Infof("Stopping server")
			return
		}

		// potentially create a ctx here
		go c.worker(c.logger, conn)
	}
}

func (c *Client) Stop() {
	c.logger.Infof("Dispatcher stopping")
	c.server.Close()
}

/*
This handles each control connection initiated by an FTP client (telnet essentially, so long running cli)

from this a context should be generated when a new go routine is spawned that initiates the data connection

if the ftp client is trying to upload/download some file


Need to figure out how to handle long running sessions based off of complex commands



         The communication path between the USER-PI and SERVER-PI for
         the exchange of commands and replies.  This connection follows
         the Telnet Protocol.
*/

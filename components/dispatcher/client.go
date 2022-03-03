package dispatcher

import (
	"bufio"
	"fmt"
	"goftp/components/logger"
	"io"
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

		go c.handleConnection(conn)
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
func (c *Client) handleConnection(conn net.Conn) {

	// start a context here

	// should for {... and parse/write back to this conn}
	// Since telent essentially starts a "long running" CLI to issue commands that
	// are generally understood by the FTP server
	c.logger.Infof("Connection recv")
	var buffer []byte
	var err error
	defer conn.Close()
	for {
		//effectively Read doesn't know when to stop from the stream input
		// we haven't specified a protocol yet, so there is no notion of knowing wh
		// idea is to read what ever we can to a buffer intermediary
		// then read from that buffer
		// read, _ := conn.Read(buffer)
		// if read > 0 {
		// 	fmt.Println(buffer)
		// 	return
		// }
		// //conn.Write([]byte("response\n"))
		// conn.Write(buffer)

		switch buffer, err = bufio.NewReader(conn).ReadBytes('\n'); err {
		case nil:
			conn.Write(buffer)
		case io.EOF:
			c.logger.Infof("Recvd EOF, Connection closed")
			return
		default:
			c.logger.Infof(fmt.Sprintf("Recvd err %s, Connection Closed", err.Error()))
			return
		}
	}

}

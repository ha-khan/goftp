package dispatcher

import (
	"bufio"
	"fmt"
	"goftp/logging"
	"io"
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

/*
This handles each control connection initiated by an FTP client (telnet essentially, so long running cli)

from this a context should generated when a new go routine is spawned that initiates the data connection

if the ftp client is trying to upload/download some file

*/
func (c *Client) handleConnection(conn net.Conn) {

	// start a context here

	// should for {... and parse/write back to this conn}
	// Since telent essentially starts a "long running" CLI to issue commands that
	// are generally understood by the FTP server
	c.log.Infof("Connection recv")
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
		case io.EOF:
			c.log.Infof("Connection closed")
			return
		default:
			c.log.Infof(fmt.Sprintf("Recv err %s", err.Error()))
			return
		}
		conn.Write(buffer)
	}

}

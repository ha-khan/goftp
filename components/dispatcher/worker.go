package dispatcher

import (
	"bufio"
	"fmt"
	"goftp/components/logger"
	"io"
	"net"
)

func worker(logger logger.Client, conn net.Conn) {
	// start a context here
	// grab logger from ctx

	// should for {... and parse/write back to this conn}
	// Since telent essentially starts a "long running" CLI to issue commands that
	// are generally understood by the FTP server
	logger.Infof("Connection recv")
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
			logger.Infof("Recvd EOF, Connection closed")
			return
		default:
			logger.Infof(fmt.Sprintf("Recvd err %s, Connection Closed", err.Error()))
			return
		}
	}
}

package worker

import (
	"bufio"
	"fmt"
	"goftp/components/logger"
	"net"
)

// worker handles the entire lifecycle management of each control connection
// initiated against the ftp server
//
// ftp is a request/response protocol and each read (request) will have its write (response)
// after synchronous execution of the handler is done
//
// ftp connections are terminated at will by the client
// need to keep track of session information, which is done by using a struct
type Worker struct {
	logger logger.Client

	// keeps track of currently logged in users
	users       map[string]string
	currentUser string
	loggedIn    bool

	// present working directory
	pwd string
}

func New(l logger.Client) *Worker {
	return &Worker{
		logger: l,
		users: map[string]string{
			"hkhan": "password",
		},
		pwd: "/usr/local/temp",
	}
}

// Start will kick off the this workers processing
// of the client initiated "control connection"
// used to issue
func (w *Worker) Start(conn net.Conn) {
	w.logger.Infof("Connection recv")
	defer conn.Close()
	defer w.logger.Infof("Closing conn")
	//
	//
	// reply to client that we're ready to start processing requests
	conn.Write(ServiceReady.Byte())
	reader := bufio.NewReader(conn)
	for {
		//
		// each FTP cmd ends with a <CRLF>
		// https://developer.mozilla.org/en-US/docs/Glossary/CRLF
		//
		//
		buffer, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println(fmt.Sprintf("%s", err.Error()))
			conn.Close()
			return
		}

		handler, req, err := w.Parse(string(buffer))
		if err != nil {
			w.logger.Infof(fmt.Sprintf("Parsing Error: %s", err.Error()))
		}

		resp, err := handler(req)
		if err != nil {
			w.logger.Infof(fmt.Sprintf("Error: %s", err.Error()))
		}

		conn.Write(resp.Byte())
		if resp == UserQuit {
			return
		}
	}
}

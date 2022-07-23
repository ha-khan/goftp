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

	// usernames -> passwords
	users    map[string]string
	loggedIn bool
	// TODO: need to keep track of current user
	// that is to be logged in. since the password
	// request will be decoupled in an other req/resp pair

	// present working directory
	pwd string
}

func NewWorker(l logger.Client) *Worker {
	return &Worker{
		logger: l,
		users: map[string]string{
			"hkhan": "password",
		},
	}
}

// Start will kick off the this workers processing
// of the client initiated control connection
func (w *Worker) Start(conn net.Conn) {
	w.logger.Infof("Connection recv")
	var buffer []byte
	var err error
	defer conn.Close()
	conn.Write(ServiceReady.Byte())

	reader := bufio.NewReader(conn)
	for {
		//
		// each FTP cmd ends with a <CRLF>
		// https://developer.mozilla.org/en-US/docs/Glossary/CRLF
		//
		//
		buffer, err = reader.ReadBytes('\n')
		if err != nil {
			fmt.Println(fmt.Sprintf("%s", err.Error()))
			conn.Close()
			return
		}

		handler, req, err := w.Parse(string(buffer))
		if err != nil {
			fmt.Println(fmt.Sprintf("%s", err.Error()))
			// TODO: need to handle this scenario
			//       and not cont
		}

		// grab response from handler also
		var resp Response
		resp, err = handler(req)
		if err != nil {
			w.logger.Infof(fmt.Sprintf("Error: %s", err.Error()))
		}

		conn.Write(resp.Byte())
	}
}

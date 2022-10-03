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
// in RFC 959 terms, this worker would be known as the Server-PI,
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

	// present working directory, clients are unable to
	// move outside of pwd, which is considered the
	// "root" directory set at initialization time
	pwd string

	// allows communication of spawned data transfer port go routine
	// this worker will feed bytes which will be sent
	// through the channel, when done the shutdown (cancel) func
	// will be invoked
	stream   chan []byte
	shutdown func()

	// signals whether STOR/RETR finished successfully or not
	done chan error
}

func New(l logger.Client) *Worker {
	return &Worker{
		logger: l,
		users: map[string]string{
			"hkhan": "password",
		},
		pwd:    "/temp",
		stream: make(chan []byte),
		done:   make(chan error),
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
			w.logger.Infof(fmt.Sprintf("Handler Error: %s", err.Error()))
		}

		// TODO: create a better approach to handle more complex FTP commands
		conn.Write(resp.Byte())
		if resp == UserQuit {
			return
		}
		if resp == StartTransfer {
			<-w.done // wait for done
			conn.Write(TransferComplete.Byte())
		}
	}
}

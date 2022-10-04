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
// execution of the handler is done
//
// ftp connections are terminated at will by the client
// need to keep track of session information, which is done by using a struct
type Worker struct {
	//
	// Info, Debug
	logger logger.Client
	//
	// keeps track of currently logged in users
	users       map[string]string
	currentUser string
	loggedIn    bool
	//
	//
	// present working directory, clients are unable to
	// move outside of pwd, which is considered the
	// "root" directory set at initialization time
	pwd string
	//
	//
	// F - File (no record structure)
	// R - Record structure
	// P - Page structure
	structure rune
	//
	//
	// S - stream
	// B - block
	// C - compress
	mode rune
	//
	//
	//
	reprType rune
	//
	//
	//
	// allows communication of spawned data transfer port go routine
	// this worker will feed bytes which will be sent
	// through the channel, when done the shutdown (cancel) func
	// will be invoked
	connection chan struct {
		socket net.Conn
		err    error
	}
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
		pwd: "/temp",
		connection: make(chan struct {
			socket net.Conn
			err    error
		}),
		shutdown:  func() {},
		done:      make(chan error),
		mode:      'S',
		structure: 'F',
	}
}

// Start will kick off the this workers processing
// of the client initiated "control connection"
// used to issue
func (w *Worker) Start(conn net.Conn) {
	w.logger.Infof("Connection recv")
	defer conn.Close()
	defer w.logger.Infof("Closing Control Connection")
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

		// TODO: certain commands require more complex interactions
		// compared to simple configuration or state information
		//
		switch conn.Write(resp.Byte()); resp {
		case UserQuit:
			return
		case StartTransfer:
			// TODO: need to handle errors
			<-w.done
			conn.Write(TransferComplete.Byte())
		default:
		}
	}
}

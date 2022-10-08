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
// ftp is stateful protocol, which means that a given request is handled
// based off of the previous requests that were handled
//
// # Thus a worker will be modeled after a finite-state machine
//
// those previous requests set specific state information
// (loggedIn, pwd, mo, stru, ty, currentOp), this tuple will be used to accept/reject a
// given Request, but checking a TransitionTable
//
// in RFC 959 terms, this worker would be known as the Server-PI,
// ftp is a request/response protocol and each read (request) will have its write (response)
// execution of the handler is done
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

	// Transfer Parameters, only accepting a subset from spec
	//
	// MODE command specifies how the bits of the data are to be transmitted
	// S - Stream
	mo rune
	//
	//
	// STRUcture and TYPE commands, are used to define the way in which the data are to be represented.
	//
	// F - File (no structure, file is considered to be a sequence of data bytes)
	// R - Record (must be accepted for "text" files (ASCII) )
	stru rune
	//
	// A - ASCII (primarily for the transfer of text files <CRLF> used to denote end of text line)
	// I - Image (data is sent as contiguous bits, which  are packed into 8-bit transfer bytes)
	ty rune

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
	// either passive
	transferComplete chan error
	currentCMD       CMD
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
		shutdown:         func() {},
		transferComplete: make(chan error),
		currentCMD:       "NONE",
		mo:               'S', // Stream
		stru:             'F', // File
		ty:               'A', // ASCII
	}
}

// Start will kick off the this workers processing
// of the client initiated control connection
//
// much of the core logic that drives the control connection is
// handled here such as errors, responses, and more complex workflows
// such as the actual transfer of bytes across the data connection
//
// errors are rarely
//
// TODO: break this into two streams, receiver cmds, and responding
func (w *Worker) Start(conn net.Conn) {
	defer func() {
		w.logger.Infof("Closing Control Connection")
		conn.Close()
	}()

	// reply to ftp client that we're ready to start processing requests
	conn.Write(ServiceReady.Byte())
	reader := bufio.NewReader(conn)
	for {
		buffer, err := reader.ReadBytes('\n')
		if err != nil {
			w.logger.Infof(fmt.Sprintf("Connection Buffer Read Error: %v", err))
			return
		}

		handler, req, err := w.Parse(string(buffer))
		if err != nil {
			// TODO: we only throw an error if we can't parse, which means we should
			// run the handler and return the response to the user
			//
			// since we failed to parse the received command, we can avoid doing
			// the transitionFunction Check
			w.logger.Infof(fmt.Sprintf("Parsing Error: %v", err))
		}

		if reject := w.RejectCMD(req); reject {
			handler = func(r *Request) (Response, error) {
				return FileActionNotTaken, nil
			}
		}

		// FIXME: We should consider errors thrown by handler as either
		// recoverable or unrecoverable
		resp, err := handler(req)
		if err != nil {
			w.logger.Infof(fmt.Sprintf("Handler Error: %v", err))
		}

		// TODO: should invoke a go routine for the writeback
		switch conn.Write(resp.Byte()); resp {
		case UserQuit:
			return
		case StartTransfer:
			if err = <-w.transferComplete; err != nil {
				w.logger.Infof(fmt.Sprintf("Transfer Error: %v", err))
				// TODO: write back that transfer failed
			} else {
				conn.Write(TransferComplete.Byte())
			}

			w.currentCMD = None
		default:
		}
	}
}

// TODO: implement this method to close any resources (channels, connections, etc)
// thus ending this Worker
func (w *Worker) Stop() {
	// gracefully shutdown worker
	// reject all subsequent commands
}

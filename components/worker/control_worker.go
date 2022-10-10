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

	currentCMD CMD
	dataWorker *DataWorker
}

func New(l logger.Client) *Worker {
	return &Worker{
		logger: l,
		users: map[string]string{
			"hkhan": "password",
		},
		pwd:        "/temp",
		currentCMD: "NONE",
		mo:         'S', // Stream
		stru:       'F', // File
		ty:         'A', // ASCII
	}
}

// Start will kick off the this workers processing
// of the client initiated control connection
//
// much of the core logic that drives the control connection is
// handled here such as errors, responses, and more complex workflows
// such as the actual transfer of bytes across the data connection
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
			w.logger.Infof(fmt.Sprintf("Parsing Error: %v", err))
		}

		if reject := w.RejectCMD(req); reject {
			handler = func(r *Request) (Response, error) {
				return FileActionNotTaken, nil
			}
		}

		resp, err := handler(req)
		if err != nil {
			w.logger.Infof(fmt.Sprintf("Handler Error: %v", err))
		}

		switch resp {
		case UserQuit:
			conn.Write(resp.Byte())
			// TODO: need to ensure that data connection is also cleaned up
			// w.dataWorker.Disconnect()
			return
		case StartTransfer:
			conn.Write(resp.Byte())
			w.dataWorker.StartTransfer(func(err error, resp Response) {
				if err != nil {
					w.logger.Infof(fmt.Sprintf("Transfer Error: %v", err))
				}

				conn.Write(resp.Byte())
				w.currentCMD = None
			})
		default:
			conn.Write(resp.Byte())
		}
	}
}

// TODO: implement this method to close any resources (channels, connections, etc)
// thus ending this Worker
func (w *Worker) Stop() {
	// gracefully shutdown worker
	// reject all subsequent commands
	if w.dataWorker != nil {
		w.dataWorker.Disconnect()
	}
}

package worker

import (
	"bufio"
	"fmt"
	"goftp/components/logger"
	"net"
)

// ControlWorker handles the entire lifecycle management of each control connection
// initiated against the ftp server
//
// ftp is stateful protocol, which means that a given request is handled
// based off of how the previous requests were handled
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
type ControlWorker struct {
	logger logger.Client

	// keeps track of currently logged in users
	users       map[string]string
	currentUser string
	loggedIn    bool

	// present working directory, clients are unable to
	// move outside of pwd, which is considered the
	// "root" directory set at initialization time
	pwd string

	transfer

	// keeps track of currently executing command, if that command is considered
	// complex ~ pasv/port/retr/...etc, and forces as specific sequence of allowable
	// commands to be called if set to a value other than 'None'
	currentCMD CMD

	// there is a 1-to-1 relation between ControlWorkers and DataWorkers
	IDataWorker
}

func NewControlWorker(l logger.Client) *ControlWorker {
	return &ControlWorker{
		logger: l,
		users: map[string]string{
			"hkhan": "password",
		},
		pwd:        "/temp",
		currentCMD: "NONE",
		transfer: transfer{
			Mode:      'S', // Stream
			Structure: 'F', // File
			Type:      'A', // ASCII
		},
		IDataWorker: NewDataWorker(
			transfer{
				Mode:      'S', // Stream
				Structure: 'F', // File
				Type:      'A', // ASCII
			},
			"/temp",
			l,
		),
	}
}

// Start will kick off the this workers processing
// of the client initiated control connection
//
// much of the core logic that drives the control connection is
// handled here such as errors, responses, and more complex workflows
// such as the actual transfer of bytes across the data connection
func (c *ControlWorker) Start(conn net.Conn) {
	defer func() {
		c.logger.Infof("Closing Control Connection")
		conn.Close()
	}()

	// reply to ftp client that we're ready to start processing requests
	conn.Write(ServiceReady.Byte())
	reader := bufio.NewReader(conn)
	for {
		buffer, err := reader.ReadBytes('\n')
		if err != nil {
			c.logger.Infof(fmt.Sprintf("Connection Buffer Read Error: %v", err))
			return
		}

		handler, req, err := c.Parse(string(buffer))
		if err != nil {
			c.logger.Infof(fmt.Sprintf("Parsing Error: %v", err))
		}

		if reject := c.RejectCMD(req); reject {
			handler = func(r *Request) (Response, error) {
				return FileActionNotTaken, nil
			}
		}

		resp, err := handler(req)
		if err != nil {
			c.logger.Infof(fmt.Sprintf("Handler Error: %v", err))
		}

		switch conn.Write(resp.Byte()); resp {
		case UserQuit:
			c.IDataWorker.Stop()
			return
		case StartTransfer:
			c.IDataWorker.Start(func(err error, resp Response) {
				if err != nil {
					c.logger.Infof(fmt.Sprintf("Transfer Error: %v", err))
				}

				conn.Write(resp.Byte())
				c.currentCMD = None
			})
		default:
			// pass through to next cmd since no "special" processing is required
		}
	}
}

// TODO: implement this method to close any resources (channels, connections, etc)
// thus ending this Worker
func (c *ControlWorker) Stop() {
	// gracefully shutdown worker
	// reject all subsequent commands
	c.IDataWorker.Stop()
}

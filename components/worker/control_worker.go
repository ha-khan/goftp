package worker

import (
	"bufio"
	"context"
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
// each worker will be modeled after a finite-state machine, primarily for requests/events
// that require a specific sequence (PASV, STOR, ..etc)
//
// those previous requests set specific state information
// (loggedIn, pwd, mo, stru, ty, currentOp), this tuple will be used to accept/reject a
// given Request, but checking a TransitionTable
//
// in RFC 959 terms, this worker would be known as the Server-PI,
// ftp is a request/response protocol and each read (request) will have its write (response)
// execution of the handler is done
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

	Transfer

	// keeps track of currently executing command, if that command is considered
	// complex ~ pasv/port/retr/...etc, and forces as specific sequence of allowable
	// commands to be called if set to a value other than 'None'
	executing CMD

	// there is a 1-to-1 relation between ControlWorkers and DataWorkers
	IDataWorker
}

func NewControlWorker(l logger.Client) *ControlWorker {
	return &ControlWorker{
		logger: l,
		users: map[string]string{
			"hkhan": "password",
		},
		pwd:         "/temp",
		executing:   "NONE",
		Transfer:    NewDefaultTransfer(),
		IDataWorker: NewDataWorker(NewDefaultTransfer(), "/temp", l),
	}
}

// Start will kick off the this workers processing
// of the client initiated control connection
//
// much of the core logic that drives the control connection is
// handled here such as errors, responses, and more complex workflows
// such as the actual transfer of bytes across the data connection
func (c *ControlWorker) Start(ctx context.Context, conn net.Conn) {
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
				c.executing = None
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

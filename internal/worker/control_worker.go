package worker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"goftp/internal/logger"
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

	dtpRespond     chan chan Response
	generalRespond chan Response

	connection net.Conn

	IExecutingState interface {
		CheckCMD(*Request) bool
		SetCMD(CMD)
		GetCMD() CMD
	}

	IDataWorker interface {
		Start(chan Response)
		Stop()
		SetTransferRequest(*Request)
		Connect(*Request) Response
	}
}

func NewControlWorker(l logger.Client, conn net.Conn) *ControlWorker {
	return &ControlWorker{
		logger: l,
		users: map[string]string{
			"hkhan": "password",
		},
		pwd:             "/temp",
		IExecutingState: NewExecutingState(),
		Transfer:        NewDefaultTransfer(),
		IDataWorker:     NewDataWorker(NewDefaultTransfer(), "/temp", l),
		connection:      conn,
		dtpRespond:      make(chan chan Response, 2),
		generalRespond:  make(chan Response, 2),
	}
}

// Start will kick off the this workers processing
// of the client initiated control connection
//
// much of the core logic that drives the control connection is
// handled here such as errors, responses, and more complex workflows
// such as the actual transfer of bytes across the data connection
func (c *ControlWorker) Receiver() {
	defer func() {
		close(c.generalRespond)
		close(c.dtpRespond)
	}()

	c.generalRespond <- ServiceReady
	for reader := bufio.NewReader(c.connection); ; {
		buffer, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				c.logger.Infof(fmt.Sprintf("Receiver: read error: %v", err))
			}

			c.generalRespond <- ForcedShutDown
			return
		}

		handler, req, err := c.Parse(string(buffer))
		if err != nil {
			c.logger.Infof(fmt.Sprintf("Receiver: parsing error: %v", err))
		}

		if reject := c.IExecutingState.CheckCMD(req); reject {
			handler = func(r *Request) (Response, error) {
				return BadSequence, nil
			}
		}

		resp, err := handler(req)
		if err != nil {
			c.logger.Infof(fmt.Sprintf("Receiver: handler error: %v", err))
		}

		switch c.generalRespond <- resp; resp {
		case UserQuit:
			return
		case StartTransfer:
			pipe := make(chan Response, 2)
			c.dtpRespond <- pipe
			c.IDataWorker.Start(pipe)
		default:
			// pass through to next cmd since no "special" processing is required
		}
	}
}

func (c *ControlWorker) Responder(ctx context.Context) {
	defer func() {
		c.IDataWorker.Stop()
		c.connection.Close()
		c.logger.Infof("Responder: closing control connection")
	}()
	for {
		select {
		case <-ctx.Done():
			if c.IExecutingState.GetCMD() != None {
				c.connection.Write(TransferAborted.Byte())
			} else {
				c.connection.Write(ServiceNotAvailable.Byte())
			}
			return
		case resp := <-c.generalRespond:
			if resp == ForcedShutDown {
				c.logger.Infof("Responder: forcing shutdown")
				return
			}
			c.connection.Write(resp.Byte())
			if resp == UserQuit {
				return
			}
		case pipe := <-c.dtpRespond:
			select {
			case <-ctx.Done():
				if c.IExecutingState.GetCMD() != None {
					c.connection.Write(TransferAborted.Byte())
				}
				return
			case resp := <-pipe:
				c.connection.Write(resp.Byte())
				c.IExecutingState.SetCMD(None)
			case resp := <-c.generalRespond:
				if resp == ForcedShutDown {
					c.logger.Infof("Received Forced Shutdown")
					return
				}
				c.connection.Write(resp.Byte())
				if resp == UserQuit {
					return
				}
			}
		}
	}
}

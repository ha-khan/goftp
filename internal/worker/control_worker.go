package worker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"goftp/internal/logger"
	"net"
)

// Each FTP request will have a corresponding handler
// paradigm is similar to how a HTTP Server would handle
// a specific HTTP Request
type Handler func(*Request) (Response, error)

// ControlWorker handles the entire lifecycle management of each control connection
// initiated against the ftp server
//
// ftp is stateful protocol, which means that a given request is handled
// based off of how the previous requests were handled
//
// each ControlWorker will be modeled after a finite-state machine, primarily for requests/events
// that require a specific sequence (PASV, STOR, ..etc)
type ControlWorker struct {
	logger logger.Client

	// keeps track of currently logged in users
	users       map[string]string
	currentUser string
	loggedIn    bool

	// channels used to send responses back to the FTP Client
	// across the Control Connection
	// buffered channels are used mainly to avoid blocking on
	// shutdown scenarios
	dtpRespond     chan chan Response
	generalRespond chan Response

	// connection with FTP Client (Control Connection)
	connection net.Conn

	// ControlWorkers can be put into a state that forces a subsequent
	// command to match a specific one, mainly for data transfers
	// also protects against malicious FTP Clients
	IExecutingState interface {
		CheckCMD(*Request) bool
		SetCMD(CMD)
		GetCMD() CMD
	}

	// There is a 1-to-1 relation with a DataWorker which handles all
	// the data transfer interactions, the ControlWorker signals to the
	// DataWorker when/what transfer should be done
	IDataWorker interface {
		Start(chan Response)
		Stop()
		SetTransferRequest(*Request)
		Connect(*Request) Response
	}

	// Data transfers can be configured and requires some "Factory" like
	// object to construct the appropriate Reader/Writer, depending on the
	// configurations set
	ITransferFactory interface {
		SetPWD(string)
		GetPWD() string
		SetStructure(rune)
		SetMode(rune)
		SetType(rune)
		GetType() rune
	}
}

func NewControlWorker(l logger.Client, conn net.Conn) *ControlWorker {
	dw := NewDataWorker(l)
	return &ControlWorker{
		logger: l,
		users: map[string]string{
			"hkhan": "password",
		},
		IExecutingState:  NewExecutingState(),
		IDataWorker:      dw,
		ITransferFactory: dw,
		connection:       conn,
		dtpRespond:       make(chan chan Response, 2),
		generalRespond:   make(chan Response, 2),
	}
}

// Receiver will kick off the this workers processing
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

// Responder multiplexes multiple channels and sends back responses to the FTP Client
// and will also do extra processing for special Responses as well as initiate a shutdown
// if the app is shutting down
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
					c.logger.Infof("Responder: Received Forced Shutdown")
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

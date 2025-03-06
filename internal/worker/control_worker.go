package worker

import (
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
	ctx    context.Context
	logger logger.Client

	// keeps track of currently logged in users
	// TODO: create dedicated struct that couples this information
	users       map[string]string
	currentUser string
	loggedIn    bool

	// connection with FTP Client (Control Connection)
	// TODO: wrap this in another object that keeps track of more information
	// control worker and data worker on not responsible for ensuring connection close
	// let them exit when they feel an error has occurred
	// create another go routine on Dispatcher that cleans up "closed" connections
	//
	controlConnection *Connection

	// ControlWorkers can be put into a state that forces a subsequent
	// command to match a specific one, mainly for data transfers
	// also protects against malicious FTP Clients
	state *State

	// There is a 1-to-1 relation with a DataWorker which handles all
	// the data transfer interactions, the ControlWorker signals to the
	// DataWorker when/what transfer should be done
	dataWorker interface {
		// orchestrates DataWorker
		Read() <-chan Response
		Start()
		Stop()
		Connect(*Request) Response

		// configures the type of transfer
		SetTransferRequest(*Request)
		SetPWD(string)
		GetPWD() string
		SetStructure(rune)
		SetMode(rune)
		SetType(rune)
		GetType() rune
	}
}

func NewControlWorker(ctx context.Context, l logger.Client, conn net.Conn) *ControlWorker {
	return &ControlWorker{
		ctx:    ctx,
		logger: l,
		users: map[string]string{
			"hkhan": "password",
		},
		state:             NewState(),
		dataWorker:        NewDataWorker(ctx, l),
		controlConnection: NewConnection(ctx, conn),
	}
}

// start this workers processing of control connection
// all write backs to the control connection happen here
func (c *ControlWorker) Start() {
	defer func() {
		c.controlConnection.Stop()
		c.dataWorker.Stop()
	}()

	c.controlConnection.Write(ServiceReady)
	for {
		var payload payload
		select {
		case <-c.ctx.Done():
			// received shutdown signal
			c.controlConnection.Write(ServiceNotAvailable)
			return
		case dataConnectionResponse := <-c.dataWorker.Read():
			c.controlConnection.Write(dataConnectionResponse)
			c.state.Set(None)
			continue
		case payload = <-c.controlConnection.Read():
			if payload.Err != nil {
				if !errors.Is(payload.Err, net.ErrClosed) {
					c.logger.Info(fmt.Sprintf("Receiver: read error: %v", payload.Err))
				}
				c.controlConnection.Write(ForcedShutDown)
				return
			}
		}

		handler, req, err := c.Parse(string(payload.Data))
		if err != nil {
			c.logger.Info(fmt.Sprintf("Receiver: parsing error: %v", err))
		}

		// TODO: move this to the parse code
		handler = c.state.Check(req, handler)
		response, err := handler(req)
		if err != nil {
			c.logger.Info(fmt.Sprintf("Receiver: handler error: %v", err))
		}

		c.controlConnection.Write(response)
		if response == UserQuit {
			// exit
			return
		}
	}
}

func (c *ControlWorker) Stop() {}

package dispatcher

import (
	"errors"
	"fmt"
	"goftp/components/logger"
	"goftp/components/worker"
	"net"
)

type Dispatcher struct {
	logger logger.Client
	server net.Listener

	// control connection port
	// usually 21
	port string

	// cancelFunc
	// TODO: store connection pool for PASV
}

// TODO: make more configurable such as TLS TCP server
func New(log logger.Client) *Dispatcher {
	return &Dispatcher{
		logger: log,
	}
}

// Start kicks off the reactor loop for each control connections initiated by some ftp client
func (d *Dispatcher) Start() {
	d.logger.Infof("Dispatcher starting up...")

	var err error
	d.server, err = net.Listen("tcp", ":2023")
	if err != nil {
		panic(err.Error())
	}

	for {
		d.logger.Infof("Dispatcher waiting for connections")

		conn, err := d.server.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				d.logger.Infof("Dispatcher shutdown complete")
				return
			}
			d.logger.Infof(fmt.Sprintf("Dispatcher connection error: %v", err))
			continue
		}

		go worker.NewControlWorker(d.logger).Start(conn)
	}
}

func (d *Dispatcher) Stop() {
	d.logger.Infof("Dispatcher shutting down...")
	// TODO:, need to keep track of all outstanding workers
	// which themselves have a connection that they are processing
	// can close gracefully or keep them alive until the client closes them
	// regardless the dispatcher needs to stop accepting new connections at a
	// minimum
	d.server.Close()
}

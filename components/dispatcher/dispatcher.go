package dispatcher

import (
	"fmt"
	"goftp/components/logger"
	"goftp/components/worker"
	"net"
)

type Dispatcher struct {
	logger logger.Client
	server net.Listener
	port   string

	// cancelFunc
}

// TODO: make more configurable such as TLS TCP server
func New(log logger.Client) *Dispatcher {
	return &Dispatcher{
		logger: log,
	}
}

// Start kicks off the reactor loop for each control connections initiated by some ftp client
func (d *Dispatcher) Start() {
	d.logger.Infof("Dispatcher starting")

	var err error
	d.server, err = net.Listen("tcp", ":2023")
	if err != nil {
		panic(err.Error())
	}

	for {
		d.logger.Infof("waiting for conn")

		conn, err := d.server.Accept()
		if err != nil {
			// write back to connection that error with server

			// TODO: need to rethink this scenario, main thread would still be blocking, should
			//       probably throw panic to kill process
			//       can invoke some cancel context passed to each worker to finish processing a
			//       request
			fmt.Println(err.Error())
			d.logger.Infof("Stopping server")

			return
		}

		go worker.New(d.logger).Start(conn)
	}
}

func (d *Dispatcher) Stop() {
	d.logger.Infof("Dispatcher stopping")
	// todo, need to keep track of all outstanding workers
	// which themselves have a connection that they are processing
	// can close gracefully or keep them alive until the client closes them
	// regardless the dispatcher needs to stop accepting new connections at a
	// minimum
	d.server.Close()
}

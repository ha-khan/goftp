package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"goftp/components/logger"
	"goftp/components/worker"
	"net"
	"sync"
	"time"
)

type Dispatcher struct {
	logger logger.Client
	server net.Listener

	// control connection port
	// usually 21
	port string

	// TODO: store connection pool for PASV
	shutdown context.CancelFunc
	wg       *sync.WaitGroup
}

// TODO: make more configurable such as TLS TCP server
func New(log logger.Client) *Dispatcher {
	return &Dispatcher{
		logger: log,
		wg:     new(sync.WaitGroup),
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

	ctx, cancel := context.WithCancel(context.Background())
	d.shutdown = cancel
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

		// TODO: pass a cancel context to each spawned worker
		// to gracefully shutdown
		d.wg.Add(1)
		go func() {
			worker.NewControlWorker(d.logger).Start(ctx, conn)
			d.wg.Done()
		}()
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
	d.shutdown()

	done := make(chan struct{}, 1)
	go func() {
		d.wg.Wait()
		done <- struct{}{}
	}()

	timeout := time.Tick(1 * time.Minute)

	select {
	case <-timeout:
		d.logger.Infof("Timeout received for shutdown, exiting")
	case <-done:
		d.logger.Infof("Graceful shutdown done, exiting")
	}
}

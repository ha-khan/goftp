package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"goftp/internal/logger"
	"goftp/internal/worker"
	"log"
	"net"
	"sync"
	"time"
)

// Dispatcher will handle all control connections initiated against the FTP Server
type Dispatcher struct {
	logger   logger.Client
	server   net.Listener
	port     string
	shutdown context.CancelFunc
	wg       *sync.WaitGroup
}

func New(log logger.Client) *Dispatcher {
	return &Dispatcher{
		logger: log,
		wg:     new(sync.WaitGroup),
	}
}

// Start kicks off the reactor loop that handles each control connections initiated by some ftp client
// a new ControlWorker instance will handle the LCM of that connection
func (d *Dispatcher) Start() {
	d.logger.Info("Dispatcher starting up...")

	var err error
	d.server, err = net.Listen("tcp", ":2023")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.shutdown = cancel
	for {
		d.logger.Info("Dispatcher waiting for connections")

		conn, err := d.server.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			d.logger.Info(fmt.Sprintf("Dispatcher connection error: %v", err))
			continue
		}

		worker := worker.NewControlWorker(ctx, d.logger, conn)
		d.wg.Add(1)
		go func() {
			worker.Start()
			d.wg.Done()
		}()
	}
}

// Stop Dispatcher thread from accepting new connections and invoke
// shutdown ctx for each subsequent worker, forces a shutdown rather than
// waiting until some transfer has completed as its a non-deterministic operation
//
// there can be future enhancements to wait for a transfer to complete in a given timeout
func (d *Dispatcher) Stop() {
	d.logger.Info("Dispatcher shutting down...")
	d.server.Close()
	d.shutdown()

	done := make(chan struct{}, 1)
	go func() {
		d.wg.Wait()
		done <- struct{}{}
	}()

	timeout := time.Tick(5 * time.Minute)
	select {
	case <-timeout:
		d.logger.Info("Timeout received for shutdown, exiting")
	case <-done:
		d.logger.Info("Shutdown done, exiting")
	}
	d.logger.Info("Dispatcher shutdown complete")
}

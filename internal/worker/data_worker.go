package worker

import (
	"fmt"
	"goftp/internal/logger"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Each DataWorker handles a single transfer request
// access to it is controlled by the ControlWorker which
// does the appropriate configuration/setup before use
type DataWorker struct {
	server net.Listener
	conn   net.Conn
	//
	// channel used to communicate with subsequent go routine
	// spawned to handler ~ Store, Retrieve, List, ... etc
	connection chan struct {
		socket net.Conn
		err    error
	}

	logger logger.Client

	host string
	port uint16
	pasv bool

	// data worker is configured to work with s specific
	// transfer request ~ Store, Retrieve, List, ... etc
	transferReq  *Request
	transferType string

	// unused at the moment, idea is the generate a r/w based off of configurations
	// to be used in Pipe(..)
	*TransferFactory
}

func NewDataWorker(logger logger.Client) *DataWorker {
	return &DataWorker{
		logger:          logger,
		TransferFactory: NewDefaultTransferFactory(),
	}
}

func (d *DataWorker) SetTransferRequest(req *Request) {
	d.transferType = req.Cmd
	d.transferReq = req
}

func (d *DataWorker) Start(resp chan Response) {
	if d.transferType == "RETR" {
		d.Pipe(resp, os.Open)
	} else if d.transferType == "STOR" {
		d.Pipe(resp, os.Create)
	}
}

// disconnect any open connections and
func (d *DataWorker) Stop() {
	d.disconnect()
}

func (d *DataWorker) Connect(req *Request) Response {
	var response Response
	if req.Cmd == "PASV" {
		d.pasv = true
		response = d.passive()
	} else {
		d.pasv = false
		response = d.active(req)
	}

	return response
}

// clean up conn
func (d *DataWorker) disconnect() {
	if d.server != nil {
		d.server.Close()
	}

	if d.conn != nil {
		d.conn.Close()
	}
}

func (d *DataWorker) Pipe(resp chan Response, file func(string) (*os.File, error)) {
	go func() {
		defer func() {
			d.disconnect()
			close(resp)
			d.logger.Info("DataWorker: Closing Data Connection")
		}()

		if d.transferReq == nil {
			resp <- SyntaxError2
			return
		}

		// TODO: eventually use TransferFactory.Create(..)
		fd, err := file("./" + d.GetPWD() + "/" + d.transferReq.Arg)
		if err != nil {
			resp <- FileNotFound
			return
		}
		defer fd.Close()

		conn := <-d.connection
		if conn.err != nil {
			resp <- CannotOpenDataConnection
			return
		}

		if conn.socket == nil {
			resp <- CannotOpenDataConnection
			return
		}

		var dst io.Writer
		var src io.Reader
		if d.transferReq.Cmd == "STOR" {
			dst, src = fd, conn.socket
		} else {
			dst, src = conn.socket, fd
		}

		_, err = io.Copy(dst, src)
		if err != nil {
			resp <- TransferAborted
			return
		}

		resp <- TransferComplete
	}()
}

func (d *DataWorker) list() {}

func (d *DataWorker) delete() {}

func (d *DataWorker) passive() Response {
	var err error
	var counter uint
retry:
	port := uint16(rand.Uint32())
	d.server, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		counter++
		if counter > 5 {
			return CannotOpenDataConnection
		}

		goto retry
	}

	err = d.server.(*net.TCPListener).SetDeadline(time.Now().Add(3 * time.Minute))
	if err != nil {
		return CannotOpenDataConnection
	}

	d.connection = make(chan struct {
		socket net.Conn
		err    error
	})
	ready := make(chan struct{})
	go func() {
		defer close(d.connection)
		var err error

		ready <- struct{}{}
		d.conn, err = d.server.Accept()

		timeout := make(chan struct{})
		defer close(timeout)
		go func() { time.Sleep(3 * time.Minute); <-timeout }()
		select {
		case d.connection <- struct {
			socket net.Conn
			err    error
		}{d.conn, err}:
		case timeout <- struct{}{}:
			d.logger.Info("DataWorker: Timout waiting for data connection to be used, shutting down")
			d.disconnect()
		}
	}()

	<-ready
	return GeneratePassiveResponse(port)
}

func (d *DataWorker) active(req *Request) Response {
	ready := make(chan error)
	d.connection = make(chan struct {
		socket net.Conn
		err    error
	})
	defer close(ready)
	go func() {
		defer close(d.connection)
		var err error

		strs := strings.Split(req.Arg, ",")
		MSB, err := strconv.Atoi(strs[4])
		if err != nil {
			ready <- err
			return
		}

		LSB, err := strconv.Atoi(strs[5])
		if err != nil {
			ready <- err
			return
		}

		port := uint16(MSB)<<8 + uint16(LSB)
		d.conn, err = net.Dial("tcp", strings.Join(strs[:4], ".")+":"+fmt.Sprintf("%d", port))
		if err != nil {
			ready <- err
			return
		}

		ready <- nil

		timeout := make(chan struct{})
		defer close(timeout)
		go func() { time.Sleep(3 * time.Minute); <-timeout }()
		select {
		case d.connection <- struct {
			socket net.Conn
			err    error
		}{d.conn, nil}:
		case timeout <- struct{}{}:
			d.logger.Info("DataWorker: Timout waiting for data connection to be used, shutting down")
			d.disconnect()
		}
	}()

	if err := <-ready; err != nil {
		return CannotOpenDataConnection
	}

	return CommandOK
}

package worker

import (
	"fmt"
	"goftp/internal/logger"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type DataWorker struct {
	// store references to both networking resources
	// to close them when needed
	server net.Listener
	conn   net.Conn
	//
	// channel used to communicate with subsequent go routine
	// spawned to handler ~ Store, Retrieve, List, ... etc
	//
	// from invocation in w.Start(...) in ClientWorker
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
		response = d.createPasv()
	} else {
		d.pasv = false
		response = d.createPort(req)
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
			d.logger.Infof("Closing Data Connection")
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

func (d *DataWorker) createPasv() Response {
	ready := make(chan error)
	d.connection = make(chan struct {
		socket net.Conn
		err    error
	})
	defer close(ready)
	go func() {
		defer close(d.connection)
		var err error

		d.server, err = net.Listen("tcp", ":2024")
		if err != nil {
			ready <- err
			return
		}

		err = d.server.(*net.TCPListener).SetDeadline(time.Now().Add(3 * time.Minute))
		if err != nil {
			ready <- err
			return
		}

		// setup successful,
		ready <- nil

		d.conn, err = d.server.Accept()
		timeout := make(chan struct{})
		defer close(timeout)
		go func() { time.Sleep(3 * time.Minute); <-timeout }()
		select {
		case d.connection <- struct {
			socket net.Conn
			err    error
		}{d.conn, nil}:
		case timeout <- struct{}{}:
			d.logger.Infof("Timout waiting for data connection to be used, shutting down")
			if d.conn != nil {
				d.conn.Close()
			}
			d.server.Close()
		}
	}()

	if err := <-ready; err != nil {
		return CannotOpenDataConnection
	}

	return PassiveMode
}

func (d *DataWorker) createPort(req *Request) Response {
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
			d.logger.Infof("Timout waiting for data connection to be used, shutting down")
			d.conn.Close()
		}
	}()

	if err := <-ready; err != nil {
		return CannotOpenDataConnection
	}

	return CommandOK
}

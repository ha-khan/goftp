package worker

import (
	"fmt"
	"goftp/internal/logger"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type DataWorker struct {
	// store references to both networking resources
	// to close them when needed
	server net.Listener
	conn   net.Conn

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

	pwd string
	Transfer
	pasv bool

	// data worker is configured to work with s specific
	// transfer request ~ Store, Retrieve, List, ... etc
	transferReq  *Request
	transferType string
}

func NewDataWorker(t Transfer, pwd string, logger logger.Client) *DataWorker {
	return &DataWorker{
		pwd:      pwd,
		logger:   logger,
		Transfer: t,
	}
}

func (d *DataWorker) SetTransferRequest(req *Request) {
	d.transferType = req.Cmd
	d.transferReq = req
}

func (d *DataWorker) Start(resp chan Response) {
	if d.transferType == "RETR" {
		d.retrieve(resp)
	} else if d.transferType == "STOR" {
		d.store(resp)
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

func (d *DataWorker) retrieve(resp chan Response) {
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

		fd, err := os.Open("./" + d.pwd + "/" + d.transferReq.Arg)
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

		_, err = io.Copy(conn.socket, fd)
		if err != nil {
			resp <- TransferAborted
			return
		}

		resp <- TransferComplete
	}()
}

func (d *DataWorker) store(resp chan Response) {
	go func() {
		defer func() {
			d.disconnect()
			close(resp)
			d.logger.Infof("Closing Data Connection")
		}()

		conn := <-d.connection
		if conn.err != nil {
			resp <- CannotOpenDataConnection
			return
		}

		fd, err := os.Create("." + d.pwd + "/" + d.transferReq.Arg)
		if err != nil {
			resp <- FileNotFound
			return
		}
		defer fd.Close()

		_, err = io.Copy(fd, conn.socket)
		if err != nil {
			resp <- TransferAborted
			return
		}

		resp <- TransferComplete
	}()
}

func (d *DataWorker) list() {

}

func (d *DataWorker) delete() {
	// TODO: need to lock file if its being used for read/store
	//
	//	delete shouldn't impact that
}

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
		if ready <- err; err != nil {
			return
		}

		// TODO: add timeout here
		d.conn, err = d.server.Accept()
		d.connection <- struct {
			socket net.Conn
			err    error
		}{
			d.conn,
			err,
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

		// TODO: add timeout here
		ready <- nil
		d.connection <- struct {
			socket net.Conn
			err    error
		}{
			d.conn,
			err,
		}
	}()

	if err := <-ready; err != nil {
		return CannotOpenDataConnection
	}

	return CommandOK
}

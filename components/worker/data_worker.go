package worker

import (
	"bufio"
	"fmt"
	"goftp/components/logger"
	"net"
	"os"
	"strconv"
	"strings"
)

type Callback func(error, Response)

type IDataWorker interface {
	Start(Callback)
	Stop()
	SetTransferRequest(*Request)
	Connect(*Request) Response
}

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
		connection: make(chan struct {
			socket net.Conn
			err    error
		}),
		pwd:      pwd,
		logger:   logger,
		Transfer: t,
	}
}

func (d *DataWorker) SetTransferRequest(req *Request) {
	d.transferType = req.Cmd
	d.transferReq = req
}

func (d *DataWorker) Start(cb Callback) {
	if d.transferType == "RETR" {
		d.retrieve(cb)
	} else if d.transferType == "STOR" {
		d.store(cb)
	}
}

// disconnect any open connections and
// close any open channels
// DataWorker is considered halted for use
// TODO: need to implement graceful close of pending transfers
// they should complete
func (d *DataWorker) Stop() {
	d.disconnect()
	close(d.connection)
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

func (d *DataWorker) retrieve(cb Callback) {
	go func() {
		defer d.disconnect()
		if d.transferReq == nil {
			cb(fmt.Errorf("Received nil transferReq"), "")
			return
		}

		fd, err := os.Open("./" + d.pwd + "/" + d.transferReq.Arg)
		if err != nil {
			cb(err, "")
			return
		}

		defer fd.Close()

		conn := <-d.connection
		if conn.err != nil {
			cb(conn.err, "")
			return
		}

		// TODO: reading/sending of bytes is based off of transfer mode
		scanner := bufio.NewScanner(fd)
		sender := bufio.NewWriter(conn.socket)
		for scanner.Scan() {
			sender.Write(append(scanner.Bytes(), []byte("\n")...))
			sender.Flush()
		}

		cb(nil, TransferComplete)
	}()
}

func (d *DataWorker) store(cb Callback) {
	go func() {
		defer d.disconnect()
		conn := <-d.connection

		fd, err := os.Create("." + d.pwd + "/" + d.transferReq.Arg)
		if err != nil {
			cb(err, "")
			return
		}
		defer fd.Close()

		diskWriter := bufio.NewWriter(fd)

		// read text into memory and then write to disk
		// FIXME: no way to know if ascii from client has new line at end
		for scanner := bufio.NewScanner(conn.socket); scanner.Scan(); {
			text := scanner.Text()
			diskWriter.WriteString(text + "\n")
			diskWriter.Flush()
		}

		cb(nil, TransferComplete)
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
	go func() {
		var err error
		d.server, err = net.Listen("tcp", ":2024")
		if ready <- err; err != nil {
			return
		}

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
	go func() {
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

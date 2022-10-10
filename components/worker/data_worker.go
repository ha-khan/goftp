package worker

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type callback func(error, Response)

type DataWorker struct {
	// embedded
	server net.Listener
	conn   net.Conn

	// channel used to communicate with subsequent ~ Store, Retrieve, List, .. etc
	connection chan struct {
		socket net.Conn
		err    error
	}

	host string
	port uint16

	pwd          string
	pasv         bool
	transferReq  *Request
	transferType string
}

func NewDataWorker(req *Request, pasv bool, pwd string) *DataWorker {
	return &DataWorker{
		pasv: pasv,
		connection: make(chan struct {
			socket net.Conn
			err    error
		}),
		pwd: pwd,
	}
}

func (d *DataWorker) retrieve(cb callback) {
	go func() {
		defer d.Disconnect()
		if d.transferReq == nil {
			cb(fmt.Errorf("Received nil transferReq"), "")
			return
		}

		fd, err := os.Open("./" + d.pwd + "/" + d.transferReq.Arg)
		if err != nil {
			cb(err, "")
			return
		}

		conn := <-d.connection
		if conn.err != nil {
			cb(conn.err, "")
			return
		}

		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			conn.socket.Write(append(scanner.Bytes(), []byte("\n")...))
		}

		cb(nil, TransferComplete)
	}()
}

func (d *DataWorker) store(cb callback) {
	go func() {
		defer d.Disconnect()
		conn := <-d.connection
		bytes, _ := io.ReadAll(conn.socket)
		fmt.Print(string(bytes))

		cb(nil, TransferComplete)
	}()
}

func (d *DataWorker) list() {

}

func (d *DataWorker) delete() {

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

func (d *DataWorker) SetTransferType(t string) error {
	if t != "RETR" && t != "STOR" {
		return fmt.Errorf("Invalid transfer type")
	}

	d.transferType = t
	return nil
}

func (d *DataWorker) SetTransferRequest(req *Request) {
	d.transferReq = req
}

func (d *DataWorker) StartTransfer(cb callback) {
	if d.transferType == "RETR" {
		d.retrieve(cb)
	} else if d.transferType == "STOR" {
		d.store(cb)
	}
}

// clean up conn
func (d *DataWorker) Disconnect() {

	if d.server != nil {
		d.server.Close()
	}

	if d.conn != nil {
		d.conn.Close()
	}

	// TODO: need to figure out how to close this
	// if quit was received right after PASV or PORT
	close(d.connection)
}

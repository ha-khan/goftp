package worker

import (
	"context"
	"fmt"
	"net"
)

type Handler func(*Request) (Response, error)

func (w Worker) checkIfLoggedIn(fn Handler) Handler {
	return func(req *Request) (Response, error) {
		if w.loggedIn {
			return fn(req)
		}

		return NotLoggedIn, fmt.Errorf("client not authenticated to run CMD")
	}
}

func (w *Worker) handleUserLogin(req *Request) (Response, error) {
	if w.loggedIn {
		return UserLoggedIn, nil
	}

	if _, ok := w.users[req.Arg]; !ok {
		return NotLoggedIn, fmt.Errorf("username: %s, not recognized", req.Arg)
	}

	// set current user for this worker
	w.currentUser = req.Arg
	return UserOkNeedPW, nil
}

func (w *Worker) handleUserPassword(req *Request) (Response, error) {
	if pw, ok := w.users[w.currentUser]; ok {
		if pw == req.Arg {
			w.loggedIn = true
			return UserLoggedIn, nil
		}
	}

	return NotLoggedIn, fmt.Errorf("incorrect password received for username %s", w.currentUser)
}

func (w Worker) handlePWD(req *Request) (Response, error) {
	return Response(fmt.Sprintf(string(DirectoryResponse), w.pwd)), nil
}

func (w Worker) handleQuit(req *Request) (Response, error) {
	return UserQuit, nil
}

func (w Worker) handleSyntaxErrorParams(req *Request) (Response, error) {
	return SyntaxError2, nil
}

func (w Worker) handleSyntaxErrorInvalidCmd(req *Request) (Response, error) {
	return SyntaxError1, nil
}

func (w Worker) handleCmdNotImplemented(req *Request) (Response, error) {
	return CmdNotImplemented, nil
}

/*
REPRESENTATION TYPE (TYPE)

		The argument specifies the representation type as described
		in the Section on Data Representation and Storage.  Several
		types take a second parameter.  The first parameter is
		denoted by a single Telnet character, as is the second
		Format parameter for ASCII and EBCDIC; the second parameter
		for local byte is a decimal integer to indicate Bytesize.
		The parameters are separated by a <SP> (Space, ASCII code
		32).

		The following codes are assigned for type:

		              \    /
		    A - ASCII |    | N - Non-print
		              |-><-| T - Telnet format effectors
		    E - EBCDIC|    | C - Carriage Control (ASA)
		              /    \
		    I - Image

		    L <byte size> - Local byte Byte size

	    The default representation type is ASCII Non-print.  If the
	    Format parameter is changed, and later just the first
	    argument is changed, Format then returns to the Non-print
	    default.
*/
func (w *Worker) handleType(req *Request) (Response, error) {
	return CommandOK, nil
}

/*
PASSIVE (PASV)

		This command requests the server-DTP to "listen" on a data
		port (which is not its default data port) and to wait for a
		connection rather than initiate one upon receipt of a
		transfer command.  The response to this command includes the
		host and port address this server is listening on.


		a 32-bit internet host address and a 16-bit TCP port address.
	    This address information is broken into 8-bit fields and the
	    value of each field is transmitted as a decimal number (in
	    character string representation).
*/
func (w *Worker) handlePassive(req *Request) (Response, error) {
	ctx, cancel := context.WithCancel(context.Background())
	w.shutdown = cancel

	ready := make(chan struct{}) // avoid possible race
	go func() {
		server, err := net.Listen("tcp", ":2024")
		if err != nil {
			panic(err.Error())
		}
		ready <- struct{}{}

		conn, err := server.Accept()
		if err != nil {
			panic(err.Error())
		}

		defer func() {
			w.logger.Infof("Closing Data Connection")
			conn.Close()
			server.Close()
		}()

		w.logger.Infof("Starting PASV")

		for {
			select {
			case <-ctx.Done():
				return
			case data := <-w.stream:
				conn.Write(data)
			}
		}
	}()

	<-ready
	return PassiveMode, nil
}

func (w *Worker) handleRetrieve(req *Request) (Response, error) {
	go func() {
		w.stream <- []byte("Hello this is a test\n")
		w.shutdown()
		w.done <- nil
	}()

	return StartTransfer, nil
}

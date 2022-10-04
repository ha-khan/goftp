package worker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
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
	ready := make(chan error)
	go func() {
		server, err := net.Listen("tcp", ":2024")
		ready <- err
		if err != nil {
			return
		}

		conn, err := server.Accept()
		w.connection <- struct {
			socket net.Conn
			err    error
		}{
			conn,
			err,
		}

		<-ctx.Done()
		w.logger.Infof("Closing Data Connection")
		conn.Close()
		server.Close()
	}()

	if err := <-ready; err != nil {
		return CannotOpenDataConnection, err
	}

	return PassiveMode, nil
}

/*
DATA PORT (PORT)

	The argument is a HOST-PORT specification for the data port
	to be used in data connection. There are defaults for both
	the user and server data ports, and under normal
	circumstances this command and its reply are not needed.  If
	this command is used, the argument is the concatenation of a
	32-bit internet host address and a 16-bit TCP port address.

	This address information is broken into 8-bit fields and the
	value of each field is transmitted as a decimal number (in
	character string representation).  The fields are separated
	by commas.  A port command would be:

	   PORT h1,h2,h3,h4,p1,p2

	where h1 is the high order 8 bits of the internet host
	address.
*/
func (w *Worker) handlePort(req *Request) (Response, error) {
	return "", nil
}

/*
will attempt to open file at PWD/req.Arg

TODO: this method needs to be a generic reader of bytes, can't
use a scanner which assumes underlying bytes are text and will
have \n to stop each scan
*/
func (w *Worker) handleRetrieve(req *Request) (Response, error) {
	go func() {
		fd, err := os.Open("./" + w.pwd + "/" + req.Arg)
		if err != nil {
			fmt.Println(err.Error())
		}

		conn := <-w.connection
		if conn.err != nil {
			w.shutdown()
			w.done <- conn.err
			return
		}

		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			conn.socket.Write(append(scanner.Bytes(), []byte("\n")...))
		}

		w.shutdown()
		w.done <- nil
	}()

	return StartTransfer, nil
}

func (w *Worker) handleStore(req *Request) (Response, error) {
	go func() {
		conn := <-w.connection

		bytes, _ := io.ReadAll(conn.socket)
		fmt.Print(string(bytes))
		w.shutdown()
		w.done <- nil
	}()

	return StartTransfer, nil
}

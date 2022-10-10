package worker

import (
	"fmt"
)

type Handler func(*Request) (Response, error)

func (w Worker) checkIfLoggedIn(fn Handler) Handler {
	return func(req *Request) (Response, error) {
		if w.loggedIn {
			return fn(req)
		}

		w.logger.Infof(fmt.Sprintf("client not authenticated to run CMD"))
		return NotLoggedIn, nil
	}
}

func (w *Worker) handleUserLogin(req *Request) (Response, error) {
	if w.loggedIn {
		return UserLoggedIn, nil
	}

	if _, ok := w.users[req.Arg]; !ok {
		w.logger.Infof(fmt.Sprintf("username: %s, not recognized", req.Arg))
		return NotLoggedIn, nil
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

	w.logger.Infof(fmt.Sprintf("incorrect password received for username %s", w.currentUser))
	return NotLoggedIn, nil
}

func (w *Worker) handleReinitialize(req *Request) (Response, error) {
	w.currentUser = ""
	w.loggedIn = false
	return Response(fmt.Sprintf(string(DirectoryResponse), w.pwd)), nil
}

func (w Worker) handleQuit(req *Request) (Response, error) {
	return UserQuit, nil
}

func (w Worker) handlePWD(req *Request) (Response, error) {
	return Response(fmt.Sprintf(string(DirectoryResponse), w.pwd)), nil
}

func (w Worker) handleNoop(req *Request) (Response, error) {
	return CommandOK, nil
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
	if len(req.Arg) != 1 {
		return SyntaxError2, nil
	}

	symbol := rune(req.Arg[0])
	if symbol != 'A' && symbol != 'I' {
		return CmdNotImplementedForParam, nil
	}

	w.ty = symbol
	return CommandOK, nil
}

/*
TRANSFER MODE (MODE)

	The argument is a single Telnet character code specifying
	the data transfer modes described in the Section on
	Transmission Modes.

	The following codes are assigned for transfer modes:

	   S - Stream
	   B - Block
	   C - Compressed

	The default transfer mode is Stream.
*/
func (w *Worker) handleMode(req *Request) (Response, error) {
	if len(req.Arg) != 1 {
		return SyntaxError2, nil
	}

	symbol := rune(req.Arg[0])
	if symbol != 'S' {
		return CmdNotImplementedForParam, nil
	}

	w.mo = symbol

	return CommandOK, nil
}

/*
FILE STRUCTURE (STRU)

	The argument is a single Telnet character code specifying
	file structure described in the Section on Data
	Representation and Storage.

	The following codes are assigned for structure:

	   F - File (no record structure)
	   R - Record structure
	   P - Page structure

	The default structure is File.
*/
func (w *Worker) handleStrucure(req *Request) (Response, error) {
	if len(req.Arg) != 1 {
		return SyntaxError2, nil
	}

	symbol := rune(req.Arg[0])
	if symbol != 'F' && symbol != 'R' {
		return CmdNotImplementedForParam, nil
	}

	if symbol == 'R' && w.ty != 'A' {
		return CmdNotImplementedForParam, nil
	}

	w.stru = symbol

	return CommandOK, nil
}

// DELE
//
//	250
//	450, 550
//	500, 501, 502, 421, 530
func (w *Worker) handleDelete(req *Request) (Response, error) {
	return TransferComplete, nil
}

//---------------------------------------------------------------------------
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

	FIXME: need to make this resilient against weird cmd sequences
*/
func (w *Worker) handlePort(req *Request) (Response, error) {
	dw := NewDataWorker(req, false, w.pwd)
	w.dataWorker = dw
	w.currentCMD = Port

	return dw.Connect(req), nil
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

	FIXME: Need to make this resilient against weird cmd sequences
*/
func (w *Worker) handlePassive(req *Request) (Response, error) {
	dw := NewDataWorker(req, true, w.pwd)
	w.dataWorker = dw
	w.currentCMD = Pasv

	return dw.Connect(req), nil
}

// RETR
//
//	125, 150
//	   (110)
//	   226, 250
//	   425, 426, 451
//	450, 550
//	500, 501, 421, 530
//
// TODO: Need to take into account the TYPE command
// if TYPE A, we can use a generic scanner, else
//
// TODO: need to coordinate, after sending StartTransfer response
// back against the control connection, we then start the retrieve
func (w *Worker) handleRetrieve(req *Request) (Response, error) {
	w.currentCMD = Retrieve
	w.dataWorker.SetTransferType("RETR")
	w.dataWorker.SetTransferRequest(req)
	return StartTransfer, nil
}

// STOR
//
//	125, 150
//	   (110)
//	   226, 250
//	   425, 426, 451, 551, 552
//	532, 450, 452, 553
//	500, 501, 421, 530
//
// TODO: need to coordinate, after sending StartTransfer response
// back against the control connection, we then start the store
func (w *Worker) handleStore(req *Request) (Response, error) {
	w.currentCMD = Store
	w.dataWorker.SetTransferType("STOR")
	w.dataWorker.SetTransferRequest(req)
	return StartTransfer, nil
}

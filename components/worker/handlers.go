package worker

import (
	"fmt"
)

type Handler func(*Request) (Response, error)

func (c ControlWorker) checkIfLoggedIn(fn Handler) Handler {
	return func(req *Request) (Response, error) {
		if c.loggedIn {
			return fn(req)
		}

		c.logger.Infof(fmt.Sprintf("client not authenticated to run CMD"))
		return NotLoggedIn, nil
	}
}

func (c *ControlWorker) handleUserLogin(req *Request) (Response, error) {
	if c.loggedIn {
		return UserLoggedIn, nil
	}

	if _, ok := c.users[req.Arg]; !ok {
		c.logger.Infof(fmt.Sprintf("username: %s, not recognized", req.Arg))
		return NotLoggedIn, nil
	}

	// set current user for this worker
	c.currentUser = req.Arg
	return UserOkNeedPW, nil
}

func (c *ControlWorker) handleUserPassword(req *Request) (Response, error) {
	if pw, ok := c.users[c.currentUser]; ok {
		if pw == req.Arg {
			c.loggedIn = true
			return UserLoggedIn, nil
		}
	}

	c.logger.Infof(fmt.Sprintf("incorrect password received for username %s", c.currentUser))
	return NotLoggedIn, nil
}

func (c *ControlWorker) handleReinitialize(req *Request) (Response, error) {
	c.currentUser = ""
	c.loggedIn = false
	return Response(fmt.Sprintf(string(DirectoryResponse), c.pwd)), nil
}

func (c ControlWorker) handleQuit(req *Request) (Response, error) {
	return UserQuit, nil
}

func (c ControlWorker) handlePWD(req *Request) (Response, error) {
	return Response(fmt.Sprintf(string(DirectoryResponse), c.pwd)), nil
}

func (c ControlWorker) handleNoop(req *Request) (Response, error) {
	return CommandOK, nil
}

func (c ControlWorker) handleSyntaxErrorParams(req *Request) (Response, error) {
	return SyntaxError2, nil
}

func (c ControlWorker) handleSyntaxErrorInvalidCmd(req *Request) (Response, error) {
	return SyntaxError1, nil
}

func (c ControlWorker) handleCmdNotImplemented(req *Request) (Response, error) {
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
func (c *ControlWorker) handleType(req *Request) (Response, error) {
	if len(req.Arg) != 1 {
		return SyntaxError2, nil
	}

	symbol := rune(req.Arg[0])
	if symbol != 'A' && symbol != 'I' {
		return CmdNotImplementedForParam, nil
	}

	c.ty = symbol
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
func (c *ControlWorker) handleMode(req *Request) (Response, error) {
	if len(req.Arg) != 1 {
		return SyntaxError2, nil
	}

	symbol := rune(req.Arg[0])
	if symbol != 'S' {
		return CmdNotImplementedForParam, nil
	}

	c.mo = symbol

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
func (c *ControlWorker) handleStrucure(req *Request) (Response, error) {
	if len(req.Arg) != 1 {
		return SyntaxError2, nil
	}

	symbol := rune(req.Arg[0])
	if symbol != 'F' && symbol != 'R' {
		return CmdNotImplementedForParam, nil
	}

	if symbol == 'R' && c.ty != 'A' {
		return CmdNotImplementedForParam, nil
	}

	c.stru = symbol

	return CommandOK, nil
}

// DELE
//
//	250
//	450, 550
//	500, 501, 502, 421, 530
func (c *ControlWorker) handleDelete(req *Request) (Response, error) {
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
func (c *ControlWorker) handlePort(req *Request) (Response, error) {
	c.currentCMD = Port
	return c.dataWorker.Connect(req), nil
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
func (c *ControlWorker) handlePassive(req *Request) (Response, error) {
	c.currentCMD = Pasv
	return c.dataWorker.Connect(req), nil
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
func (c *ControlWorker) handleRetrieve(req *Request) (Response, error) {
	c.currentCMD = Retrieve
	c.dataWorker.SetTransferRequest(req)

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
func (w *ControlWorker) handleStore(req *Request) (Response, error) {
	w.currentCMD = Store
	w.dataWorker.SetTransferRequest(req)

	return StartTransfer, nil
}

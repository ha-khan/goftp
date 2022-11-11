package worker

import (
	"fmt"
)

type Handler func(*Request) (Response, error)

func (c ControlWorker) handlePWD(req *Request) (Response, error) {
	return Response(fmt.Sprintf(string(DirectoryResponse), c.pwd)), nil
}

func (c ControlWorker) handleNoop(req *Request) (Response, error) {
	return CommandOK, nil
}

// Transfer Parameters, only accepting a subset from spec
type Transfer struct {
	// MODE command specifies how the bits of the data are to be transmitted
	// S - Stream
	Mode rune
	//
	//
	// STRUcture and TYPE commands, are used to define the way in which the data are to be represented.
	//
	// F - File (no structure, file is considered to be a sequence of data bytes)
	// R - Record (must be accepted for "text" files (ASCII) )
	Structure rune
	//
	//
	// A - ASCII (primarily for the transfer of text files <CRLF> used to denote end of text line)
	// I - Image (data is sent as contiguous bits, which  are packed into 8-bit transfer bytes)
	Type rune
}

func NewDefaultTransfer() Transfer {
	return Transfer{
		Mode:      'S', // Stream
		Structure: 'F', // File
		Type:      'A', // ASCII
	}
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

	c.Type = symbol
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

	c.Mode = symbol
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

	if symbol == 'R' && c.Type != 'A' {
		return CmdNotImplementedForParam, nil
	}

	c.Structure = symbol
	return CommandOK, nil
}

//---------------------------------------------------------------------------

// DELE
//
//	250
//	450, 550
//	500, 501, 502, 421, 530
func (c *ControlWorker) handleDelete(req *Request) (Response, error) {
	c.executing = Delete
	return TransferComplete, nil
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
func (c *ControlWorker) handlePort(req *Request) (Response, error) {
	c.executing = Port
	return c.IDataWorker.Connect(req), nil
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
func (c *ControlWorker) handlePassive(req *Request) (Response, error) {
	c.executing = Pasv
	return c.IDataWorker.Connect(req), nil
}

// RETR
//
//	125, 150
//	   (110)
//	   226, 250
//	   425, 426, 451
//	450, 550
//	500, 501, 421, 530
func (c *ControlWorker) handleRetrieve(req *Request) (Response, error) {
	c.executing = Retrieve
	c.IDataWorker.SetTransferRequest(req)

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
func (w *ControlWorker) handleStore(req *Request) (Response, error) {
	w.executing = Store
	w.IDataWorker.SetTransferRequest(req)

	return StartTransfer, nil
}

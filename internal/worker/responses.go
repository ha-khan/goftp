package worker

import "fmt"

type Response string

func (r Response) Byte() []byte {
	return append([]byte(r), []byte(CRLF)...)
}

// TODO: eventually take in host name
func GeneratePassiveResponse(port uint16) Response {
	var MSB uint16
	var LSB uint16

	LSB = port & uint16(0x00FF)
	MSB = (port >> 8) & uint16(0x00FF)

	return Response(fmt.Sprintf("227 Entering Passive Mode (127,0,0,1,%d,%d)", MSB, LSB))
}

const (
	CRLF Response = "\r\n"
)

// custom, will not be sent back to ftp client used to control
// shutdown in case of EOF or other control connection read issues
const (
	ForcedShutDown Response = "Control Connection Read Issue"
)

// 100s
const (
	StartTransfer      Response = "125 Data connection already open; transfer starting"
	FileOKOpenDataConn Response = "150 File status okay; about to open data connection"
)

// 200s
const (
	CommandOK         Response = "200 Command okay"
	ServiceReady      Response = "220 Service Ready"
	UserQuit          Response = "221 Service closing control connection"
	UserLoggedIn      Response = "230 User logged in, proceed"
	TransferComplete  Response = "250 Requested file action okay, completed"
	DirectoryResponse Response = "257 \"%s\""
)

// 300s
const (
	UserOkNeedPW Response = "331 User name okay, need password"
)

// 400s
const (
	CannotOpenDataConnection Response = "425 Can't open data connection"
	ServiceNotAvailable      Response = "421 Service not available, closing control connection"
	TransferAborted          Response = "426 Connection closed; transfer aborted"
	FileActionNotTaken       Response = "450 Requested file action not taken"
)

// 500s
const (
	SyntaxError1              Response = "500 Syntax error, command unrecognized"
	SyntaxError2              Response = "501 Syntax error in parameters or arguments"
	CmdNotImplemented         Response = "502 Command not implemented"
	BadSequence               Response = "503 Bad sequence of commands"
	CmdNotImplementedForParam Response = "504 Command not implemented for that parameter"
	NotLoggedIn               Response = "530 Not logged in"
	FileNotFound              Response = "550 Requested action not taken"
)

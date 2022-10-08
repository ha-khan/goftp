package worker

type Response string

func (r Response) Byte() []byte {
	return append([]byte(r), []byte(CRLF)...)
}

const (
	CRLF Response = "\r\n"
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
	PassiveMode       Response = "227 Entering Passive Mode (127,0,0,1,7,232)" // 127.0.0.1:2024
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
	FileActionNotTaken                = "450 Requested file action not taken"
)

// 500s
const (
	SyntaxError1              Response = "500 Syntax error, command unrecognized"
	SyntaxError2              Response = "501 Syntax error in parameters or arguments"
	CmdNotImplemented         Response = "502 Command not implemented"
	CmdNotImplementedForParam Response = "504 Command not implemented for that parameter"
	NotLoggedIn               Response = "530 Not logged in"
)

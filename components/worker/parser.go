package worker

import (
	"fmt"
	"strings"
)

type Request struct {
	Cmd string
	Arg string
}

func (r *Request) String() string {
	return fmt.Sprintf("%s %s", r.Cmd, r.Arg)
}

func (w *Worker) Parse(request string) (Handler, *Request, error) {
	var req *Request
	switch str := strings.Split(request, " "); len(str) {
	case 2:
		req = &Request{
			Cmd: strings.ToUpper(string(str[0])),
			Arg: string(str[1][:len(str[1])-2]),
		}
	case 1:
		req = &Request{
			Cmd: string(str[0][:len(str[0])-2]),
		}
	default:
		return w.handleSyntaxErrorParams, req, fmt.Errorf("Unable to parse request")
	}

	w.logger.Infof(req.String())

	var handler Handler
	switch req.Cmd {
	case "USER":
		return w.handleUserLogin, req, nil
	case "PASS":
		return w.handleUserPassword, req, nil
	case "PWD":
		handler = w.handlePWD
	case "TYPE":
		handler = w.handleType
	case "MODE":
		handler = w.handleMode
	case "PASV":
		handler = w.handlePassive
	case "PORT":
		handler = w.handlePort
	case "STOR":
		handler = w.handleStore
	case "RETR":
		handler = w.handleRetrieve
	case "DELE":
		return nil, nil, nil
	case "NOOP":
		handler = w.handleNoop
	case "QUIT":
		return w.handleQuit, req, nil
	case "LIST", "ACCT", "CWD", "CDUP", "SMNT", "REIN", "HELP",
		"STRU", "STOU", "APPE", "ALLO", "REST", "RNFR", "RNTO",
		"ABOR", "RMD", "MKD", "NLST", "SITE", "SYST", "STAT":
		handler = w.handleCmdNotImplemented
	default:
		return w.handleSyntaxErrorInvalidCmd, req, fmt.Errorf("Invalid CMD: %s", req.Cmd)
	}

	return w.checkIfLoggedIn(handler), req, nil
}

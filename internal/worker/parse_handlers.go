package worker

import (
	"bytes"
	"fmt"
	"regexp"
)

var pattern *regexp.Regexp

func init() {
	// potentially match against CMDs, Args
	pattern = regexp.MustCompile("\r\n")
}

type Request struct {
	Cmd string
	Arg string
}

func (r *Request) String() string {
	return fmt.Sprintf("%s %s", r.Cmd, r.Arg)
}

func (c *ControlWorker) Parse(request []byte) (Handler, *Request, error) {
	var req *Request
	if !pattern.Match(request) {
		return c.handleSyntaxErrorParams, req, fmt.Errorf("request format is incorrect")
	}

	parsed := bytes.Split(request, []byte(" "))
	switch len(parsed) {
	case 2:
		req = &Request{
			Cmd: string(bytes.ToUpper(parsed[0])),
			Arg: string(parsed[1][:len(parsed[1])-2]),
		}
	case 1:
		req = &Request{
			Cmd: string(parsed[0][:len(parsed[0])-2]),
		}
	default:
		return c.handleSyntaxErrorParams, &Request{}, fmt.Errorf("unable to parse request")
	}

	c.logger.Info(req.String())

	var handler Handler
	switch req.Cmd {
	case "USER":
		return c.handleUserLogin, req, nil
	case "PASS":
		return c.handleUserPassword, req, nil
	case "PWD":
		handler = c.handlePWD
	case "TYPE":
		handler = c.handleType
	case "MODE":
		handler = c.handleMode
	case "PASV":
		handler = c.handlePassive
	case "PORT":
		handler = c.handlePort
	case "STOR":
		handler = c.handleStore
	case "RETR":
		handler = c.handleRetrieve
	case "NOOP":
		handler = c.handleNoop
	case "QUIT":
		return c.handleQuit, req, nil
	case "LIST", "ACCT", "CWD", "CDUP", "SMNT", "REIN", "HELP",
		"STRU", "STOU", "APPE", "ALLO", "REST", "RNFR", "RNTO",
		"ABOR", "RMD", "MKD", "NLST", "SITE", "SYST", "STAT", "DELE", "EPSV":
		return c.handleCmdNotImplemented, req, fmt.Errorf("CMD Not Implementd: %v", req.Cmd)
	default:
		return c.handleSyntaxErrorInvalidCmd, req, fmt.Errorf("invalid CMD: %s", req.Cmd)
	}

	return c.checkIfLoggedIn(handler), req, nil
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

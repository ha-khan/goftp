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
			Cmd: string(str[0]),
			Arg: string(str[1][:len(str[1])-2]),
		}
	case 1:
		req = &Request{
			Cmd: string(str[0][:len(str[0])-2]),
		}
	default:
		return nil, nil, fmt.Errorf("Unable to parse request")

	}

	// find appropriate handler
	var handler Handler
	switch req.Cmd {
	case "USER":
		return w.handleUserLogin, req, nil
	case "PWD":
		handler = nil
	default:
		return nil, req, fmt.Errorf("Invaled CMD: %s", req.Cmd)
	}

	return w.checkIfLoggedIn(handler), nil, nil
}

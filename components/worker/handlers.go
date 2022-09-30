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

		return NotLoggedIn, fmt.Errorf("Error: client not authenticated to run CMD")
	}
}

func (w *Worker) handleUserLogin(req *Request) (Response, error) {
	if w.loggedIn {
		return UserLoggedIn, nil
	}

	if _, ok := w.users[req.Arg]; !ok {
		return NotLoggedIn, fmt.Errorf("Username: %s, not recognized", req.Arg)
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

func (w *Worker) handlePWD(req *Request) (Response, error) {
	return Response(fmt.Sprintf(string(DirectoryResponse), w.pwd)), nil
}

func (w *Worker) handleQuit(req *Request) (Response, error) {
	return UserQuit, nil
}

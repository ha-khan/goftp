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

	// TODO: need to keep track of current user login workflow
	//       username/password are separate req/resp pairs
	return UserOkNeedPW, nil
}

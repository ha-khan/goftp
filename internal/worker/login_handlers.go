package worker

import "fmt"

func (c ControlWorker) checkIfLoggedIn(fn Handler) Handler {
	return func(req *Request) (Response, error) {
		if c.loggedIn {
			return fn(req)
		}

		c.logger.Info(fmt.Sprintf("client not authenticated to run CMD"))
		return NotLoggedIn, nil
	}
}

func (c *ControlWorker) handleUserLogin(req *Request) (Response, error) {
	if c.loggedIn {
		return UserLoggedIn, nil
	}

	if _, ok := c.users[req.Arg]; !ok {
		c.logger.Info(fmt.Sprintf("username: %s, not recognized", req.Arg))
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

	c.logger.Info(fmt.Sprintf("incorrect password received for username %s", c.currentUser))
	return NotLoggedIn, nil
}

func (c *ControlWorker) handleReinitialize(req *Request) (Response, error) {
	c.currentUser = ""
	c.loggedIn = false
	return Response(fmt.Sprintf(string(DirectoryResponse), c.DataWorker.GetPWD())), nil
}

func (c ControlWorker) handleQuit(req *Request) (Response, error) {
	return UserQuit, nil
}

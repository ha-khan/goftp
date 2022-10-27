package worker

type CMD string

const (
	None     CMD = "NONE"
	Store    CMD = "STOR"
	Retrieve CMD = "RETR"
	Delete   CMD = "DELE"
	Port     CMD = "PORT"
	Pasv     CMD = "PASV"
)

// If a command appears here, that implies that it
// should be rejected
var baseReject = map[CMD]any{
	Retrieve: nil,
	Delete:   nil,
	Store:    nil,
	Pasv:     nil,
	Port:     nil,
}

// currentCMD -> requestedCMD --> Reject ~ true | false
//
//		NONE -> Delete -------------------------> NONE
//	       \                                       ^
//		    \                                     /
//		     v                                   /
//		     (PORT | PASV) -> (Store | Retrieve)
var table = map[CMD]map[CMD]any{
	None: {
		Retrieve: nil,
		Store:    nil,
	},
	Store:    baseReject,
	Retrieve: baseReject,
	Delete:   baseReject,
	Pasv: {
		Pasv:   nil,
		Port:   nil,
		Delete: nil,
	},
	Port: {
		Port:   nil,
		Pasv:   nil,
		Delete: nil,
	},
}

// check whether the given request + current state will allow for handler to
// be invoked
//
// forces a sequence when Port/Pasv is invoked, signaling that the next
// operation should be some kind of data transfer (Store/Retrieve), which would
// eventually return the worker to an "idle" state.
//
// configuration and other lcm commands are however still accepted,
func (c *ControlWorker) RejectCMD(requested *Request) bool {
	_, reject := table[c.executing][CMD(requested.Cmd)]

	return reject
}

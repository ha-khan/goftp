package worker

import "sync"

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

type ExecutingState struct {
	// current executing state
	cmd CMD

	stateTable map[CMD]map[CMD]any

	mutex sync.Locker
}

func NewExecutingState() *ExecutingState {
	return &ExecutingState{
		cmd:        None,
		stateTable: table,
		mutex:      new(sync.Mutex),
	}
}

// check whether the given request + current state will allow for handler to
// be invoked
//
// forces a sequence when Port/Pasv is invoked, signaling that the next
// operation should be some kind of data transfer (Store/Retrieve), which would
// eventually return the worker to an "idle" state.
//
// configuration and other lcm commands are however still accepted,
func (e *ExecutingState) CheckCMD(requested *Request) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	_, reject := table[e.cmd][CMD(requested.Cmd)]

	return reject
}

func (e *ExecutingState) SetCMD(cmd CMD) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.cmd = cmd
}

func (e *ExecutingState) GetCMD() CMD {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.cmd
}

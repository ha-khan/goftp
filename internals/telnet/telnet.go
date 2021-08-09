package telnet

import (
	"sync"
)

type handlers struct {
	*sync.RWMutex
	cache map[string]func() error
}

// func newHandlers()

// var h handlers

// func init() {
// 	h = &handlers{cach}
// }

// type TelnetRequest struct {
// 	IAC, CmdCode, OptCode byte
// 	HandleRequest         func() error
// }

// func EvaluateTCPConnToTelnet(conn net.Conn) *TelnetRequest {

// 	return nil
// }

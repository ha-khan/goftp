package telnet

import "fmt"

type TelnetRequest struct {
	IAC     byte
	CmdCode byte
	OptCode []byte
	Handle  func() error
}

func Parse(request []byte) (handler *TelnetRequest, err error) {
	handler = &TelnetRequest{
		IAC:     request[0],
		CmdCode: request[1],
		OptCode: request[2:],
	}

	fmt.Print(handler)

	switch str := fmt.Sprintf("%s%s", string(handler.IAC), string(handler.CmdCode)); str {
	case "ab":
		handler.Handle = func() error { return nil }
	default:
		fmt.Print(str)
		return nil, fmt.Errorf("invaled IAC")
	}
	return
}

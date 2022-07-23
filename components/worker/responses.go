package worker

type Response string

func (r Response) Byte() []byte {
	return []byte(r)
}

const (
	CRLF         Response = "\r\n"
	ServiceReady Response = "220 Service Ready" + CRLF
	UserLoggedIn Response = "230 User logged in, proceed" + CRLF
	NotLoggedIn  Response = "530 Not logged in" + CRLF
	UserOkNeedPW Response = "331 User name okay, need password" + CRLF
)

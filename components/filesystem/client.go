package filesystem

import "io"

type Client interface {
	io.ReadWriter
	List(string) string
}

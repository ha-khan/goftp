package worker

import (
	"bufio"
	"context"
	"net"
)

type Connection struct {
	ctx context.Context

	conn net.Conn
	read interface {
		ReadBytes(byte) ([]byte, error)
	}
	write interface {
		WriteString(string) (int, error)
		Flush() error
	}

	pipe chan payload
}

type payload struct {
	Data []byte
	Err  error
}

func NewConnection(ctx context.Context, conn net.Conn) *Connection {
	c := &Connection{
		ctx:   ctx,
		conn:  conn,
		read:  bufio.NewReader(conn),
		write: bufio.NewWriter(conn),
		pipe:  make(chan payload),
	}

	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				buffer, err := c.read.ReadBytes('\n')
				c.pipe <- payload{
					Data: buffer,
					Err:  err,
				}

				if err != nil {
					return
				}
			}
		}
	}()

	return c
}

func (c *Connection) Stop() {
	c.conn.Close()
}

func (c Connection) Write(response Response) {
	c.write.WriteString(response.String())
	c.write.Flush()
}

func (c *Connection) Read() <-chan payload {
	return c.pipe
}

package logger

import (
	"log"
	"os"
	"sync"
)

var once sync.Once
var stdClient *stdStreamClient

func NewStdStreamClient() *stdStreamClient {
	once.Do(func() {
		stdClient = &stdStreamClient{
			Logger: log.New(os.Stdout, "goftp", 1),
		}
	})

	return stdClient
}

// writes to standard streams ~ stdout, stderr
type stdStreamClient struct {
	*log.Logger
}

func (s *stdStreamClient) SetLevel() {
	s.Infof("Setting level to ...")
}

// Infof ...
func (s *stdStreamClient) Infof(msg string) {
	s.Println(msg)
}

package logger

import (
	"log/slog"
	"os"
	"sync"
)

var once sync.Once
var stdClient *slog.Logger

func NewStdStreamClient() Client {
	once.Do(func() {
		stdClient = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	})

	return stdClient
}

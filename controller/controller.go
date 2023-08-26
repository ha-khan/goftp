package controller

import (
	"goftp/internal/dispatcher"
	"goftp/internal/logger"
	"sync"
)

var once sync.Once
var goFtp *GoFTP

type GoFTP struct {
	logger     logger.Client
	dispatcher *dispatcher.Dispatcher
}

func BasicGoFTP() *GoFTP {
	once.Do(func() {
		logger := logger.NewStdStreamClient()
		goFtp = &GoFTP{
			logger:     logger,
			dispatcher: dispatcher.New(logger),
		}
	})

	return goFtp
}

// Start kicks off more Go routines which are expected to be running until the lifetime of the process
func (g *GoFTP) Start() {
	g.logger.Info("Starting up GoFTP...")
	go g.dispatcher.Start()
}

// Stop will shutdown the service
func (g *GoFTP) Stop() {
	g.logger.Info("Shutting down GoFTP...")
	g.dispatcher.Stop()
	g.logger.Info("GoFTP shutdown complete, exiting")
}

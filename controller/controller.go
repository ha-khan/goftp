package controller

import (
	"goftp/components/dispatcher"
	"goftp/components/logger"
	"sync"
)

var once sync.Once
var goFtp *GoFTP

// GoFTP is the singleton instance of this
type GoFTP struct {
	logger     logger.Client
	dispatcher *dispatcher.Dispatcher
	// TODO: eventually add a DataConnectionManager here
	// which allows us to reuse open Ports
}

// BasicGoFTP returns a basic GOFTP instance
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
	g.logger.Infof("Starting up GoFTP...")
	go g.dispatcher.Start()
}

// Stop stops all the internal components that
func (g *GoFTP) Stop() {
	g.logger.Infof("Gracefully shutting down GoFTP...")
	g.dispatcher.Stop()
	g.logger.Infof("GoFTP graceful shutdown complete, exiting")
}

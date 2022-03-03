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
	dispatcher *dispatcher.Client
	// TODO: create a worker *worker.Client which handles connections from dispatcher
	//
}

// NewBasicGoFTP returns a basic GOFTP instance
func NewBasicGoFTP() *GoFTP {
	once.Do(func() {
		logger := logger.NewStdStreamClient()
		goFtp = &GoFTP{
			logger:     logger,
			dispatcher: dispatcher.NewClient(logger),
		}
	})

	return goFtp
}

// Start kicks off more Go routines which are expected to be running until the lifetime of the process
func (g *GoFTP) Start() {
	g.logger.Infof("Starting GoFTP")
	go g.dispatcher.Start()
}

// Stop stops all the internal components that
func (g *GoFTP) Stop() {
	g.logger.Infof("Stopping GoFTP")
	g.dispatcher.Stop()
}

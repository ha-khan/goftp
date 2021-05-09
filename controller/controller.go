package controller

import (
	"goftp/internals/dispatcher"
	"goftp/logging"
	"sync"
)

var once sync.Once
var goFtp *GoFTP

// GoFTP is the singleton instance of this
type GoFTP struct {
	log *logging.Client

	dispatcher *dispatcher.Client
}

// NewBasicGoFTP returns a basic GOFTP instance
func NewBasicGoFTP() *GoFTP {
	once.Do(func() {
		var log *logging.Client
		log = &logging.Client{}
		goFtp = &GoFTP{log: log, dispatcher: dispatcher.NewClient(log)}
	})
	return goFtp
}

// Start kicks off more Go routines which are expected to be running until the lifetime of the process
func (g *GoFTP) Start() error {
	g.log.Infof("Starting GoFTP")
	go g.dispatcher.Start()
	return nil
}

// Stop stops
func (g *GoFTP) Stop() {
	g.log.Infof("Stopping GoFTP")
	g.dispatcher.Stop()
}

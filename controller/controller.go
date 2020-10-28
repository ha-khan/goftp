package controller

import (
	"goftp/logging"
)

// GoFTP is the singleton instance of this
type GoFTP struct {
	*logging.Client
}

// NewBasicGoFTP returns a basic GOFTP instance
func NewBasicGoFTP() *GoFTP {
	return &GoFTP{&logging.Client{}}
}

// Start starts
func (g *GoFTP) Start() error {
	g.Client.Infof("Starting GoFTP")
	return nil
}

// Stop stops
func (g *GoFTP) Stop() error {
	g.Client.Infof("Stopping GoFTP")
	return nil
}

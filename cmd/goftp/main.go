package main

import (
	"goftp/controller"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var stop = make(chan os.Signal, 1)
	var goftp = controller.BasicGoFTP()
	// TODO: figure out a good approach to configuring internal components
	// port         = flag.String("port", "2023", "specify TCP port to expose ftp server on")

	go goftp.Start()
	signal.Notify(stop, syscall.SIGQUIT, os.Interrupt)

	<-stop
	goftp.Stop()
}

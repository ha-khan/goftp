package main

import (
	"goftp/controller"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	var (
		stop         chan os.Signal
		appSingleton *controller.GoFTP
		// TODO: figure out a good approach to configuring internal components
		// port         = flag.String("port", "2023", "specify TCP port to expose ftp server on")
	)

	appSingleton = controller.NewBasicGoFTP()

	go appSingleton.Start()

	stop = make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGQUIT, os.Interrupt)
	<-stop
	appSingleton.Stop()
}

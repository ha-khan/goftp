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
	)

	appSingleton = controller.NewBasicGoFTP()

	go appSingleton.Start()

	stop = make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGQUIT, os.Interrupt)
	<-stop
	appSingleton.Stop()
}

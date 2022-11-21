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

	go goftp.Start()
	signal.Notify(stop, syscall.SIGQUIT, os.Interrupt)

	<-stop
	goftp.Stop()
}

package main

import (
	"goftp/internal/controller"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var stop = make(chan os.Signal, 1)
	var goftp = controller.NewGoFTP()

	go goftp.Start()
	signal.Notify(stop, syscall.SIGQUIT, os.Interrupt)

	<-stop
	goftp.Stop()
}

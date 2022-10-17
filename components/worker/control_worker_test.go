package worker

import (
	"bufio"
	"fmt"
	"goftp/components/logger"
	"net"
	"testing"
)

// TODO: add fuzzing here and more complex chained interations
// this should test end-to-end
// TODO: use https://pkg.go.dev/net#Pipe to mock net.Conn without
// using any networking resources
func Test_Login(t *testing.T) {
	client, server := net.Pipe()
	worker := NewControlWorker(logger.NewStdStreamClient())

	go worker.Start(server)

	scanner := bufio.NewScanner(client)
	bufWriter := bufio.NewWriter(client)
	scanner.Scan()
	fmt.Println(scanner.Text())
	bufWriter.WriteString("PWD\r\n")
	bufWriter.Flush()
	scanner.Scan()
	fmt.Println(scanner.Text())
}

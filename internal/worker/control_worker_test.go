package worker

import (
	"bufio"
	"context"
	"goftp/internal/logger"
	"net"
	"testing"
)

func Test_Start_And_Shutdown(t *testing.T) {
	client, server := net.Pipe()
	worker := NewControlWorker(logger.NewStdStreamClient(), server)

	go worker.Start(context.TODO())

	scanner := bufio.NewScanner(client)
	writer := bufio.NewWriter(client)

	scanner.Scan()
	if resp := scanner.Text(); resp != string(ServiceReady) {
		t.Errorf("Expected: %s, but got %s", string(ServiceReady), resp)
		return

	}

	writer.WriteString("QUIT\r\n")
	writer.Flush()

	scanner.Scan()
	if resp := scanner.Text(); resp != string(UserQuit) {
		t.Errorf("Expected: %s, but got %s", string(ServiceReady), resp)
		return
	}
}

func Test_User_Login(t *testing.T) {
	client, server := net.Pipe()
	worker := NewControlWorker(logger.NewStdStreamClient(), server)

	go worker.Start(context.TODO())

	scanner := bufio.NewScanner(client)
	writer := bufio.NewWriter(client)

	scanner.Scan()
	if resp := scanner.Text(); resp != string(ServiceReady) {
		t.Errorf("Expected: %s, but got %s", string(ServiceReady), resp)
		return

	}

	writer.WriteString("USER hkhan\r\n")
	writer.Flush()
	scanner.Scan()
	if resp := scanner.Text(); resp != string(UserOkNeedPW) {
		t.Errorf("Expected: %s, but got %s", string(UserOkNeedPW), resp)
		return
	}

	writer.WriteString("PASS password\r\n")
	writer.Flush()
	scanner.Scan()
	if resp := scanner.Text(); resp != string(UserLoggedIn) {
		t.Errorf("Expected: %s, but got %s", string(UserLoggedIn), resp)
		return
	}

	writer.WriteString("QUIT\r\n")
	writer.Flush()
	scanner.Scan()
	if resp := scanner.Text(); resp != string(UserQuit) {
		t.Errorf("Expected: %s, but got %s", string(ServiceReady), resp)
		return
	}
}

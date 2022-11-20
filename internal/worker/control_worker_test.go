package worker

import (
	"bufio"
	"context"
	"goftp/internal/logger"
	"net"
	"testing"
)

func Test_Start_And_Quit(t *testing.T) {
	client, server := net.Pipe()
	worker := NewControlWorker(logger.NewStdStreamClient(), server)

	go worker.Receiver()
	go worker.Responder(context.TODO())

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
		t.Errorf("Expected: %s, but got %s", string(UserQuit), resp)
		return
	}
}

func Test_Start_And_OS_Shutdown(t *testing.T) {
	client, server := net.Pipe()
	worker := NewControlWorker(logger.NewStdStreamClient(), server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Receiver()
	go worker.Responder(ctx)

	scanner := bufio.NewScanner(client)
	scanner.Scan()
	if resp := scanner.Text(); resp != string(ServiceReady) {
		t.Errorf("Expected: %s, but got %s", string(ServiceReady), resp)
		return
	}
	cancel()

	scanner.Scan()
	if resp := scanner.Text(); resp != string(ServiceNotAvailable) {
		t.Errorf("Expected: %s, but got %s", string(ServiceNotAvailable), resp)
		return
	}
}

func Test_User_Login(t *testing.T) {
	client, server := net.Pipe()
	worker := NewControlWorker(logger.NewStdStreamClient(), server)

	go worker.Receiver()
	go worker.Responder(context.TODO())

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
		t.Errorf("Expected: %s, but got %s", string(UserQuit), resp)
		return
	}
}

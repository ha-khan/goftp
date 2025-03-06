package worker

import (
	"context"
	"goftp/internal/logger"
	"net"
	"strings"
	"testing"
)

// TODO: add fuzzing

func Test_Parse_Format_Error(t *testing.T) {
	_, server := net.Pipe()
	w := NewControlWorker(context.Background(), logger.NewStdStreamClient(), server)

	handler, req, err := w.Parse("abcd efg nhijk lmnop\r\n")
	if err == nil {
		t.Errorf("Expected not nil error")
	}

	resp, err := handler(req)
	if err != nil {
		t.Errorf("Expected nil error, but got %v", err)
	}

	if resp != SyntaxError2 {
		t.Errorf("Expected Response: %s, but got %s", SyntaxError2, resp)
	}
}

func Test_Parse_CMD_Not_Implemented(t *testing.T) {
	_, server := net.Pipe()
	w := NewControlWorker(context.Background(), logger.NewStdStreamClient(), server)
	w.loggedIn = true

	// if the cmd is recognized by RFC 959, but we're not implementing it,
	// that shouldn't return an error
	handler, req, err := w.Parse("SMNT\r\n")
	if err == nil {
		t.Errorf("Expected not nil error")
	}

	resp, err := handler(req)
	if err != nil {
		t.Errorf("Expected nil error, but got %v", err)
	}

	if resp != CmdNotImplemented {
		t.Errorf("Expected Response: %s, but got %s", CmdNotImplemented, resp)
	}

}

func Test_Parse_Invalid_Cmd(t *testing.T) {
	_, server := net.Pipe()
	w := NewControlWorker(context.Background(), logger.NewStdStreamClient(), server)

	handler, req, err := w.Parse("EPSV\r\n")
	if err == nil {
		t.Errorf("Expected not nil error")
	}

	if !strings.Contains(err.Error(), "invalid CMD:") {
		t.Errorf("Expected error message to contain 'Invalid CMD:'")
	}

	// we shouldn't expect the handler to error out as that should be
	resp, err := handler(req)
	if err != nil {
		t.Errorf("Expected nil error, but got %v", err)
	}

	if resp != SyntaxError1 {
		t.Errorf("Expected Response: %s, but got %s", SyntaxError1, resp)
	}
}

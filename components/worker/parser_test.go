package worker

import (
	"goftp/components/logger"
	"testing"
)

func Test_Parse(t *testing.T) {
	w := New(logger.NewStdStreamClient())
	_, _, err := w.Parse("USER hkhan\r\n")
	if err != nil {
		t.Errorf("Expected nil")
	}
}

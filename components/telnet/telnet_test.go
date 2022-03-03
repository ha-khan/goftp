package telnet

import "testing"

func Test_Parse(t *testing.T) {
	handler, err := Parse([]byte(`abc`))
	err = handler.Handle()

	if err != nil {
		t.Errorf("Expected nil")
	}
}

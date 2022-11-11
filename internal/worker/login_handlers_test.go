package worker

import (
	"goftp/internal/logger"
	"testing"
)

// table-driven tests for individual handlers
// https://go.dev/blog/subtests
var accessControlTestCases = []struct {
	TestName         string
	Command          string
	MutationFunc     func(w *ControlWorker)
	HandlerErrCheck  func(e error, t *testing.T)
	HandlerRespValue Response
}{
	{
		TestName:         "Test_User_Success",
		Command:          "USER hkhan\r\n",
		MutationFunc:     noMutation,
		HandlerErrCheck:  expectNilErr,
		HandlerRespValue: UserOkNeedPW,
	},
	{
		TestName:         "Test_User_Already_Logged_In",
		Command:          "USER hkhan\r\n",
		MutationFunc:     setUserLoggedIn,
		HandlerErrCheck:  expectNilErr,
		HandlerRespValue: UserLoggedIn,
	},
	{
		TestName:         "Test_User_Username_Not_Recognized",
		Command:          "USER hakhan\r\n",
		MutationFunc:     noMutation,
		HandlerErrCheck:  expectNilErr,
		HandlerRespValue: NotLoggedIn,
	},
	{
		TestName: "Test_User_Password_Incorrect",
		Command:  "PASS password123\r\n",
		MutationFunc: func(w *ControlWorker) {
			w.currentUser = "hkhan"
		},
		HandlerErrCheck:  expectNilErr,
		HandlerRespValue: NotLoggedIn,
	},
	{
		TestName: "Test_User_Password_Incorrect",
		Command:  "PASS password\r\n",
		MutationFunc: func(w *ControlWorker) {
			w.currentUser = "hkhan"
		},
		HandlerErrCheck:  expectNilErr,
		HandlerRespValue: UserLoggedIn,
	},
}

func expectNilErr(e error, t *testing.T) {
	if e != nil {
		t.Errorf("Expected nil error, but got %v", e)
	}
}

func noMutation(w *ControlWorker) { return }

func setUserLoggedIn(w *ControlWorker) {
	w.loggedIn = true
	return
}

func TestDriver(t *testing.T) {
	for _, testcase := range accessControlTestCases {
		t.Run(testcase.TestName, func(t *testing.T) {
			w := NewControlWorker(logger.NewStdStreamClient())
			testcase.MutationFunc(w)
			handler, req, err := w.Parse(testcase.Command)
			if err != nil {
				t.Errorf("Expected nil error from Parse, but got %v", err)
			}

			resp, err := handler(req)
			testcase.HandlerErrCheck(err, t)
			if resp != testcase.HandlerRespValue {
				t.Errorf("Expected Response: %s, but got %s", testcase.HandlerRespValue, resp)
			}
		})
	}
}

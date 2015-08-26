package logrus_fluent

import (
	"bytes"
	"net"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
)

var data chan string

const (
	testHOST = "localhost"
)

func TestLogEntryMessageReceivedWithTagAndMessage(t *testing.T) {
	f := logrus.Fields{
		"message": "message!",
		"tag":     "debug.test",
		"value":   "data",
	}
	result := testLog(t, f, "MyMessage1")

	switch {
	case !strings.Contains(result, "\x94\xaadebug.test\xd2"):
		t.Errorf("message did not contain tag")
	case !strings.Contains(result, "value\xa4data"):
		t.Errorf("message did not contain value")
	case !strings.Contains(result, "\xa7message\xa8message!"):
		t.Errorf("message did not contain message")
	}
}

func TestLogEntryMessageReceivedWithMessage(t *testing.T) {
	f := logrus.Fields{
		"message": "message!",
		"value":   "data",
	}
	result := testLog(t, f, "MyMessage2")

	switch {
	case !strings.Contains(result, "\xaaMyMessage2\xd2"):
		t.Errorf("message did not contain tag from entry.Message")
	case !strings.Contains(result, "value\xa4data"):
		t.Errorf("message did not contain value")
	case !strings.Contains(result, "\xa7message\xa8message!"):
		t.Errorf("message did not contain message")
	}
}

func TestLogEntryMessageReceivedWithTag(t *testing.T) {
	f := logrus.Fields{
		"tag":   "debug.test",
		"value": "data",
	}
	result := testLog(t, f, "MyMessage3")

	switch {
	case !strings.Contains(result, "\x94\xaadebug.test\xd2"):
		t.Errorf("message did not contain tag")
	case !strings.Contains(result, "value\xa4data"):
		t.Errorf("message did not contain value")
	case !strings.Contains(result, "\xa7message\xaaMyMessage3"):
		t.Errorf("message did not contain entry.Message")
	}
}

func TestLogEntryMessageReceived(t *testing.T) {
	f := logrus.Fields{
		"value": "data",
	}
	result := testLog(t, f, "MyMessage4")

	switch {
	case !strings.Contains(result, "value\xa4data"):
		t.Errorf("message did not contain value")
	case !strings.Contains(result, "\xaaMyMessage4"):
		t.Errorf("message did not contain entry.Message")
	}
}

func testLog(t *testing.T, f logrus.Fields, message string) string {
	data = make(chan string, 1)
	port := startMockServer(t)
	hook := NewHook(testHOST, port)
	logger := logrus.New()
	logger.Hooks.Add(hook)

	logger.WithFields(f).Error(message)

	return <-data
}

func startMockServer(t *testing.T) int {
	l, err := net.Listen("tcp", testHOST+":0")
	if err != nil {
		t.Errorf("Error listening: %s", err.Error())
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				t.Errorf("Error accepting: %s", err.Error())
			}
			go handleRequest(conn, l)
		}
	}()
	return l.Addr().(*net.TCPAddr).Port
}

func handleRequest(conn net.Conn, l net.Listener) {
	bf := new(bytes.Buffer)
	bf.ReadFrom(conn)
	conn.Close()
	data <- bf.String()
}

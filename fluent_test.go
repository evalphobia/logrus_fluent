package logrus_fluent

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
)

var (
	// used for persistent mock server
	data     = make(chan string)
	mockPort int
)

const (
	defaultLoopCount = 5 // assertion count
	defaultStaticTag = "STATIC_TAG"
	testHOST         = "localhost"
)

func TestNew(t *testing.T) {
	_, port := newMockServer(t, nil)
	hook, err := New(testHOST, port)
	switch {
	case err != nil:
		t.Errorf("error on New: %s", err.Error())
	case hook == nil:
		t.Errorf("hook should not be nil")
	case len(hook.levels) != len(defaultLevels):
		t.Errorf("hook.levels should be defaultLevels")
	}
}

func TestNewHook(t *testing.T) {
	const testPort = -1
	hook := NewHook(testHOST, testPort)
	switch {
	case hook == nil:
		t.Errorf("hook should not be nil")
	case hook.host != testHOST:
		t.Errorf("hook.host should be %s, but %s", testHOST, hook.host)
	case hook.port != testPort:
		t.Errorf("hook.port should be %d, but %d", testPort, hook.port)
	case len(hook.levels) != len(defaultLevels):
		t.Errorf("hook.levels should be defaultLevels")
	}
}

func TestLevels(t *testing.T) {
	hook := FluentHook{}

	levels := hook.Levels()
	if levels != nil {
		t.Errorf("hook.Levels() should be nil, but %v", levels)
	}

	hook.levels = []logrus.Level{logrus.WarnLevel}
	levels = hook.Levels()
	switch {
	case levels == nil:
		t.Errorf("hook.Levels() should not be nil")
	case len(levels) != 1:
		t.Errorf("hook.Levels() should have 1 length")
	case levels[0] != logrus.WarnLevel:
		t.Errorf("hook.Levels() should have logrus.WarnLevel")
	}
}

func TestSetLevels(t *testing.T) {
	hook := FluentHook{}

	levels := hook.levels
	if levels != nil {
		t.Errorf("hook.levels should be nil, but %v", levels)
	}

	hook.SetLevels([]logrus.Level{logrus.WarnLevel})
	levels = hook.levels
	switch {
	case levels == nil:
		t.Errorf("hook.levels should not be nil")
	case len(levels) != 1:
		t.Errorf("hook.levels should have 1 length")
	case levels[0] != logrus.WarnLevel:
		t.Errorf("hook.levels should have logrus.WarnLevel")
	}

	hook.SetLevels(nil)
	levels = hook.levels
	if levels != nil {
		t.Errorf("hook.levels should be nil, but %v", levels)
	}
}

func TestTag(t *testing.T) {
	hook := FluentHook{}

	tag := hook.Tag()
	if tag != "" {
		t.Errorf("hook.Tag() should be empty, but %s", tag)
	}

	defaultTag := defaultStaticTag
	hook.tag = &defaultTag
	tag = hook.Tag()
	switch {
	case tag == "":
		t.Errorf("hook.Tag() should not be empty")
	case tag != defaultTag:
		t.Errorf("hook.Tag() should be %s, but %s", defaultTag, tag)
	}
}

func TestSetTag(t *testing.T) {
	hook := FluentHook{}

	tag := hook.tag
	if tag != nil {
		t.Errorf("hook.tag should be nil, but %s", *tag)
	}

	hook.SetTag(defaultStaticTag)
	tag = hook.tag
	switch {
	case tag == nil:
		t.Errorf("hook.tag should not be nil")
	case *tag != defaultStaticTag:
		t.Errorf("hook.tag should be %s, but %s", defaultStaticTag, *tag)
	}
}

func TestLogEntryMessageReceived(t *testing.T) {
	f := logrus.Fields{
		"value": "data",
	}

	assertion := func(result string) {
		switch {
		case !strings.Contains(result, "value\xa4data"):
			t.Errorf("message did not contain value")
		case !strings.Contains(result, "\xaaMyMessage1"):
			t.Errorf("message did not contain entry.Message")
		}
	}
	assertLogHook(t, f, "MyMessage1", assertion)

}

func TestLogEntryMessageReceivedWithTag(t *testing.T) {
	f := logrus.Fields{
		"tag":   "debug.test",
		"value": "data",
	}

	assertion := func(result string) {
		switch {
		case !strings.Contains(result, "\x94\xaadebug.test\xd2"):
			t.Errorf("message did not contain tag")
		case !strings.Contains(result, "value\xa4data"):
			t.Errorf("message did not contain value")
		case !strings.Contains(result, "\xa7message\xaaMyMessage2"):
			t.Errorf("message did not contain entry.Message")
		}
	}
	assertLogHook(t, f, "MyMessage2", assertion)
}

func TestLogEntryMessageReceivedWithMessage(t *testing.T) {
	fmt.Printf("TestLogEntryMessageReceivedWithMessage...\n")

	f := logrus.Fields{
		"message": "message!",
		"value":   "data",
	}

	assertion := func(result string) {
		switch {
		case !strings.Contains(result, "\xaaMyMessage3\xd2"):
			t.Errorf("message did not contain tag from entry.Message")
		case !strings.Contains(result, "value\xa4data"):
			t.Errorf("message did not contain value")
		case !strings.Contains(result, "\xa7message\xa8message!"):
			t.Errorf("message did not contain message")
		}
	}
	assertLogHook(t, f, "MyMessage3", assertion)
}

func TestLogEntryMessageReceivedWithTagAndMessage(t *testing.T) {
	f := logrus.Fields{
		"message": "message!",
		"tag":     "debug.test",
		"value":   "data",
	}

	assertion := func(result string) {
		switch {
		case !strings.Contains(result, "\x94\xaadebug.test\xd2"):
			t.Errorf("message did not contain tag")
		case !strings.Contains(result, "value\xa4data"):
			t.Errorf("message did not contain value")
		case !strings.Contains(result, "\xa7message\xa8message!"):
			t.Errorf("message did not contain message")
		}
	}
	assertLogHook(t, f, "MyMessage4", assertion)
}

func TestLogEntryWithStaticTag(t *testing.T) {
	f := logrus.Fields{
		"tag":   "something",
		"value": "data",
	}

	assertion := func(result string) {
		switch {
		case !strings.Contains(result, defaultStaticTag):
			t.Errorf("message did not contain the correct, static tag")
		case !strings.Contains(result, "something"):
			t.Errorf("message did not contain the tag field")
		}
	}
	assertLogHookWithStaticTag(t, f, "MyMessage5", assertion)
}

func assertLogHook(t *testing.T, f logrus.Fields, message string, assertFunc func(string)) {
	assertLogMessage(t, f, message, "", assertFunc)
}

func assertLogHookWithStaticTag(t *testing.T, f logrus.Fields, message string, assertFunc func(string)) {
	assertLogMessage(t, f, message, defaultStaticTag, assertFunc)
}

func assertLogMessage(t *testing.T, f logrus.Fields, message string, tag string, assertFunc func(string)) {
	// assert brand new logger
	{
		localData := make(chan string)
		_, port := newMockServer(t, localData)
		hook := NewHook(testHOST, port)
		if tag != "" {
			hook.SetTag(tag)
		}
		logger := logrus.New()
		logger.Hooks.Add(hook)

		for i := 0; i < defaultLoopCount; i++ {
			logger.WithFields(f).Error(message)
			assertFunc(<-localData)
		}
	}

	// assert persistent logger
	{
		port := getOrCreateMockServer(t, data)
		hook, err := New(testHOST, port)
		if err != nil {
			t.Errorf("Error on NewHookWithLogger: %s", err.Error())
		}
		if tag != "" {
			hook.SetTag(tag)
		}

		logger := logrus.New()
		logger.Hooks.Add(hook)

		for i := 0; i < defaultLoopCount; i++ {
			logger.WithFields(f).Error(message)
			assertFunc(<-data)
		}
	}
}

func getOrCreateMockServer(t *testing.T, data chan string) int {
	if mockPort == 0 {
		_, mockPort = newMockServer(t, data)
	}
	return mockPort
}

func newMockServer(t *testing.T, data chan string) (net.Listener, int) {
	l, err := net.Listen("tcp", testHOST+":0")
	if err != nil {
		t.Errorf("Error listening: %s", err.Error())
	}

	count := 0
	go func() {
		for {
			count++
			conn, err := l.Accept()
			if err != nil {
				t.Errorf("Error accepting: %s", err.Error())
			}

			go handleRequest(conn, l, data)
			if count == defaultLoopCount {
				l.Close()
				return
			}
		}
	}()
	return l, l.Addr().(*net.TCPAddr).Port
}

func handleRequest(conn net.Conn, l net.Listener, data chan string) {
	r := bufio.NewReader(conn)
	for {
		b := make([]byte, 1<<10) // Read 1KB at a time
		_, err := r.Read(b)
		if err == io.EOF {
			continue
		} else if err != nil {
			fmt.Printf("Error reading from connection: %s", err)
		}
		data <- string(b)
	}
}

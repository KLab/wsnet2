package log_test

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"wsnet2/log"
)

func TestLevel(t *testing.T) {
	log.SetLevel(log.INFO)
	old := log.SetLevel(log.ERROR)
	if old != log.INFO {
		t.Fatalf("old level = %v, wants %v", old, log.INFO)
	}
	if l := log.CurrentLevel(); l != log.ERROR {
		t.Fatalf("current level = %v, wants %v", l, log.ERROR)
	}
}

func TestDebugf(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetWriter(buf)
	log.SetLevel(log.ALL)

	_, _, line, _ := runtime.Caller(0)
	log.Debugf("debug message %d", 1)
	wants := fmt.Sprintf("log_test.go:%d: [DEBUG] debug message 1", line+1)
	if s := buf.String(); !strings.Contains(s, wants) {
		t.Fatalf("output = \"%v\", must contains \"%v\"", s, wants)
	}
}

func TestInfof(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetWriter(buf)
	log.SetLevel(log.ALL)

	_, _, line, _ := runtime.Caller(0)
	log.Infof("info message %d", 2)
	wants := fmt.Sprintf("log_test.go:%d: [INFO] info message 2", line+1)
	if s := buf.String(); !strings.Contains(s, wants) {
		t.Fatalf("output = \"%v\", must contains \"%v\"", s, wants)
	}
}

func TestErrorf(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetWriter(buf)
	log.SetLevel(log.ALL)

	_, _, line, _ := runtime.Caller(0)
	log.Errorf("error message %d", 3)
	wants := fmt.Sprintf("log_test.go:%d: [ERROR] error message 3", line+1)
	if s := buf.String(); !strings.Contains(s, wants) {
		t.Fatalf("output = \"%v\", must contains \"%v\"", s, wants)
	}
}

func TestLogLevel(t *testing.T) {
	var s string
	buf := bytes.NewBuffer(nil)
	log.SetWriter(buf)

	log.SetLevel(log.ALL)
	log.Debugf("debug")
	log.Infof("info")
	log.Errorf("error")
	s = buf.String()
	if !strings.Contains(s, "[DEBUG]") {
		t.Fatalf("output must contains \"[DEBUG]\": \"%s\"", s)
	}
	if !strings.Contains(s, "[INFO]") {
		t.Fatalf("output must contains \"[INFO]\": \"%s\"", s)
	}
	if !strings.Contains(s, "[ERROR]") {
		t.Fatalf("output  must contains \"[ERROR]\": \"%s\"", s)
	}

	buf.Reset()
	log.SetLevel(log.DEBUG)
	log.Debugf("debug")
	log.Infof("info")
	log.Errorf("error")
	s = buf.String()
	if !strings.Contains(s, "[DEBUG]") {
		t.Fatalf("output must contains \"[DEBUG]\": \"%s\"", s)
	}
	if !strings.Contains(s, "[INFO]") {
		t.Fatalf("output must contains \"[INFO]\": \"%s\"", s)
	}
	if !strings.Contains(s, "[ERROR]") {
		t.Fatalf("output  must contains \"[ERROR]\": \"%s\"", s)
	}

	buf.Reset()
	log.SetLevel(log.INFO)
	log.Debugf("debug")
	log.Infof("info")
	log.Errorf("error")
	s = buf.String()
	if strings.Contains(s, "[DEBUG]") {
		t.Fatalf("output must not contains \"[DEBUG]\": \"%s\"", s)
	}
	if !strings.Contains(s, "[INFO]") {
		t.Fatalf("output must contains \"[INFO]\": \"%s\"", s)
	}
	if !strings.Contains(s, "[ERROR]") {
		t.Fatalf("output  must contains \"[ERROR]\": \"%s\"", s)
	}

	buf.Reset()
	log.SetLevel(log.ERROR)
	log.Debugf("debug")
	log.Infof("info")
	log.Errorf("error")
	s = buf.String()
	if strings.Contains(s, "[DEBUG]") {
		t.Fatalf("output must not contains \"[DEBUG]\": \"%s\"", s)
	}
	if strings.Contains(s, "[INFO]") {
		t.Fatalf("output must not contains \"[INFO]\": \"%s\"", s)
	}
	if !strings.Contains(s, "[ERROR]") {
		t.Fatalf("output  must contains \"[ERROR]\": \"%s\"", s)
	}

	buf.Reset()
	log.SetLevel(log.NOLOG)
	log.Debugf("debug")
	log.Infof("info")
	log.Errorf("error")
	s = buf.String()
	if len(s) > 0 {
		t.Fatalf("output must empty: %s", s)
	}
}

func TestGet(t *testing.T) {
	tests := []log.Level{log.NOLOG, log.ERROR, log.INFO, log.DEBUG, log.ALL}
	for _, l := range tests {
		logger := log.Get(l)
		if lv := logger.Level(); lv != l {
			t.Fatalf("Get(%v).Level()=%v, wants %v", l, lv, l)
		}
	}
}

func TestLogger(t *testing.T) {
	var s string
	buf := bytes.NewBuffer(nil)
	log.SetWriter(buf)

	format := map[bool]string{
		true:  "output must contains \"%s\": \"%s\"",
		false: "output must not contains \"%s\": \"%s\"",
	}

	tests := []struct {
		logger   log.Logger
		hasDebug bool
		hasInfo  bool
		hasError bool
	}{
		{log.Get(log.ALL), true, true, true},
		{log.Get(log.DEBUG), true, true, true},
		{log.Get(log.INFO), false, true, true},
		{log.Get(log.ERROR), false, false, true},
		{log.Get(log.NOLOG), false, false, false},
	}
	for _, test := range tests {
		buf.Reset()
		test.logger.Debugf("debug")
		test.logger.Infof("info")
		test.logger.Errorf("error")
		s = buf.String()
		if strings.Contains(s, "debug") != test.hasDebug {
			t.Fatalf(format[test.hasDebug], "debug", s)
		}
		if strings.Contains(s, "info") != test.hasInfo {
			t.Fatalf(format[test.hasInfo], "info", s)
		}
		if strings.Contains(s, "error") != test.hasError {
			t.Fatalf(format[test.hasError], "error", s)
		}
	}
}

func TestStringer(t *testing.T) {
	if s, w := log.ALL.String(), "ALL"; s != w {
		t.Fatalf("string \"%v\" wants \"%v\"", s, w)
	}
	if s, w := log.DEBUG.String(), "DEBUG"; s != w {
		t.Fatalf("string \"%v\" wants \"%v\"", s, w)
	}
	if s, w := log.INFO.String(), "INFO"; s != w {
		t.Fatalf("string \"%v\" wants \"%v\"", s, w)
	}
	if s, w := log.ERROR.String(), "ERROR"; s != w {
		t.Fatalf("string \"%v\" wants \"%v\"", s, w)
	}
	if s, w := log.NOLOG.String(), "NOLOG"; s != w {
		t.Fatalf("string \"%v\" wants \"%v\"", s, w)
	}
}

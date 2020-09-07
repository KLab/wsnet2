package log_test

import (
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

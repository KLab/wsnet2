package game

import (
	"reflect"
	"testing"
)

type EvTest int

var _ Event = EvTest(0)

func TestWriteRead(t *testing.T) {
	buf := NewEvBuf(5)

	evs := []Event{EvTest(0), EvTest(1), EvTest(2)}
	cseq := len(evs)
	for _, ev := range evs {
		if e := buf.Write(ev); e != nil {
			t.Fatalf("Write(%v) error: %v", ev, e)
		}
	}
	r, seq := buf.Read()
	if !reflect.DeepEqual(r, evs) || seq != cseq {
		t.Fatalf("Read (%v, %v), wants (%v, %v)", r, seq, evs, cseq)
	}

	r, seq = buf.Read()
	if len(r) != 0 || seq != cseq {
		t.Fatalf("Read (%v, %v), wants ([], %v)", r, seq, cseq)
	}

	evs = []Event{EvTest(3), EvTest(4), EvTest(5), EvTest(6), EvTest(7)}
	cseq += len(evs)
	for _, m := range evs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}
	r, seq = buf.Read()
	if !reflect.DeepEqual(r, evs) || seq != cseq {
		t.Fatalf("Read (%v, %v), wants (%v, %v)", r, seq, evs, cseq)
	}
}

func TestEvBufOverFlow(t *testing.T) {
	buf := NewEvBuf(2)
	ev := EvTest(0)

	if e := buf.Write(ev); e != nil {
		t.Fatalf("Write error: %v", e)
	}
	if e := buf.Write(ev); e != nil {
		t.Fatalf("Write error: %v", e)
	}
	if e := buf.Write(ev); e == nil {
		t.Fatalf("Write must error")
	}
}

func TestRewind(t *testing.T) {
	buf := NewEvBuf(5)

	evs := []Event{EvTest(1), EvTest(2), EvTest(3), EvTest(4)}
	for _, m := range evs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}
	buf.Read()
	evs = []Event{EvTest(5), EvTest(6), EvTest(7)}
	for _, m := range evs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}

	if e := buf.Rewind(3); e != nil {
		t.Fatalf("Rewind(3) error: %v", e)
	}

	r, seq := buf.Read()
	wants := []Event{EvTest(4), EvTest(5), EvTest(6), EvTest(7)}
	cseq := 7
	if !reflect.DeepEqual(r, wants) || seq != cseq {
		t.Fatalf("Read (%v, %v), wants (%v, %v)", r, seq, wants, cseq)
	}

	if e := buf.Rewind(1); e == nil {
		t.Fatalf("Rewind(1) must error")
	}
}

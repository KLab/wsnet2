package game

import (
	"reflect"
	"testing"
)

type EvTest int

var _ Event = EvTest(0)

func (EvTest) event() {}

func TestWriteRead(t *testing.T) {
	buf := NewEvBuf(5)

	evs := []Event{EvTest(0), EvTest(1), EvTest(2)}
	cseq := len(evs)
	for _, ev := range evs {
		if e := buf.Write(ev); e != nil {
			t.Fatalf("Write(%v) error: %v", ev, e)
		}
	}
	r, seq, err := buf.Read(0)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if !reflect.DeepEqual(r, evs) || seq != cseq {
		t.Fatalf("Read (%v, %v), wants (%v, %v)", r, seq, evs, cseq)
	}

	r, seq, err = buf.Read(seq)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
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
	r, seq, err = buf.Read(seq)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
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

func TestReadWithRewind(t *testing.T) {
	buf := NewEvBuf(5)

	evs := []Event{EvTest(1), EvTest(2), EvTest(3), EvTest(4)}
	for _, m := range evs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}
	buf.Read(0)
	evs = []Event{EvTest(5), EvTest(6), EvTest(7)}
	for _, m := range evs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}

	r, seq, e := buf.Read(3)
	if e != nil {
		t.Fatalf("Read(3) error: %v", e)
	}
	wants := []Event{EvTest(4), EvTest(5), EvTest(6), EvTest(7)}
	cseq := 7
	if !reflect.DeepEqual(r, wants) || seq != cseq {
		t.Fatalf("Read(3) (%v, %v), wants (%v, %v)", r, seq, wants, cseq)
	}

	if _, _, e := buf.Read(2); e == nil {
		t.Fatalf("Read(2) must error")
	}
}

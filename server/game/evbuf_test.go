package game

import (
	"reflect"
	"testing"

	"wsnet2/binary"
)

func TestWriteRead(t *testing.T) {
	buf := NewEvBuf(5)

	evs := []*binary.Event{
		{Type: 0},
		{Type: 1},
		{Type: 2},
	}

	for _, ev := range evs {
		if e := buf.Write(ev); e != nil {
			t.Fatalf("Write(%v) error: %v", ev, e)
		}
	}
	seq := 0
	r, err := buf.Read(0)
	if err != nil {
		t.Fatalf("Read(0) error: %v", err)
	}
	if !reflect.DeepEqual(r, evs) {
		t.Fatalf("Read(%v) %v, wants %v", seq, r, evs)
	}
	seq += len(r)

	r, err = buf.Read(seq)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if len(r) != 0 {
		t.Fatalf("Read(%v) %v, wants []", seq, r)
	}
	seq += len(r)

	evs = []*binary.Event{
		{Type: 3},
		{Type: 4},
		{Type: 5},
		{Type: 6},
		{Type: 7},
	}
	for _, m := range evs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}
	r, err = buf.Read(seq)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if !reflect.DeepEqual(r, evs) {
		t.Fatalf("Read(%v) %v, wants %v", seq, r, evs)
	}
}

func TestEvBufOverFlow(t *testing.T) {
	buf := NewEvBuf(2)
	ev := &binary.Event{Type: 0}

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

	evs := []*binary.Event{
		{Type: 1},
		{Type: 2},
		{Type: 3},
		{Type: 4},
	}
	for _, m := range evs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}
	buf.Read(0)

	evs = []*binary.Event{
		{Type: 5},
		{Type: 6},
		{Type: 7},
	}
	for _, m := range evs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}

	r, e := buf.Read(3)
	if e != nil {
		t.Fatalf("Read(3) error: %v", e)
	}
	wants := []*binary.Event{
		{Type: 4},
		{Type: 5},
		{Type: 6},
		{Type: 7},
	}
	if !reflect.DeepEqual(r, wants) {
		t.Fatalf("Read(3) %v, wants %v", r, wants)
	}

	if _, e := buf.Read(2); e == nil {
		t.Fatalf("Read(2) must error")
	}
}

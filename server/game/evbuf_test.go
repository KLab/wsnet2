package game

import (
	"reflect"
	"testing"

	"wsnet2/binary"
)

func TestWriteRead(t *testing.T) {
	buf := NewEvBuf(5)

	evs := []*binary.RegularEvent{
		binary.NewRegularEvent(0, nil),
		binary.NewRegularEvent(1, nil),
		binary.NewRegularEvent(2, nil),
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

	evs = []*binary.RegularEvent{
		binary.NewRegularEvent(3, nil),
		binary.NewRegularEvent(4, nil),
		binary.NewRegularEvent(5, nil),
		binary.NewRegularEvent(6, nil),
		binary.NewRegularEvent(7, nil),
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
	ev := binary.NewRegularEvent(0, nil)

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

	evs := []*binary.RegularEvent{
		binary.NewRegularEvent(1, nil),
		binary.NewRegularEvent(2, nil),
		binary.NewRegularEvent(3, nil),
		binary.NewRegularEvent(4, nil),
	}
	for _, m := range evs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}
	buf.Read(0)

	evs = []*binary.RegularEvent{
		binary.NewRegularEvent(5, nil),
		binary.NewRegularEvent(6, nil),
		binary.NewRegularEvent(7, nil),
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
	wants := []*binary.RegularEvent{
		binary.NewRegularEvent(4, nil),
		binary.NewRegularEvent(5, nil),
		binary.NewRegularEvent(6, nil),
		binary.NewRegularEvent(7, nil),
	}
	if !reflect.DeepEqual(r, wants) {
		t.Fatalf("Read(3) %v, wants %v", r, wants)
	}

	if _, e := buf.Read(2); e == nil {
		t.Fatalf("Read(2) must error")
	}
}

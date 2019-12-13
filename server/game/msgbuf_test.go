package game

import (
	"reflect"
	"testing"
)

func TestWriteRead(t *testing.T) {
	buf := NewMsgBuf(5)

	msgs := []Msg{0, 1, 2}
	for _, m := range msgs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}
	r := buf.Read()
	if !reflect.DeepEqual(r, msgs) {
		t.Fatalf("Read %v, wants %v", r, msgs)
	}

	r = buf.Read()
	msgs = []Msg{}
	if !reflect.DeepEqual(r, msgs) {
		t.Fatalf("Read %v, wants %v", r, msgs)
	}

	msgs = []Msg{3, 4, 5, 6, 7}
	for _, m := range msgs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}
	r = buf.Read()
	if !reflect.DeepEqual(r, msgs) {
		t.Fatalf("Read %v, wants %v", r, msgs)
	}

}

func TestMsgBufOverFlow(t *testing.T) {
	buf := NewMsgBuf(2)
	msg := Msg(struct{}{})

	if e := buf.Write(msg); e != nil {
		t.Fatalf("Write error: %v", e)
	}
	if e := buf.Write(msg); e != nil {
		t.Fatalf("Write error: %v", e)
	}
	if e := buf.Write(msg); e == nil {
		t.Fatalf("Write must error")
	}
}

func TestRewind(t *testing.T) {
	buf := NewMsgBuf(5)

	msgs := []Msg{1, 2, 3, 4}
	for _, m := range msgs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}
	buf.Read()
	msgs = []Msg{5, 6, 7}
	for _, m := range msgs {
		if e := buf.Write(m); e != nil {
			t.Fatalf("Write(%v) error: %v", m, e)
		}
	}

	if e := buf.Rewind(3); e != nil {
		t.Fatalf("Rewind(3) error: %v", e)
	}

	r := buf.Read()
	wants := []Msg{4, 5, 6, 7}
	if !reflect.DeepEqual(r, wants) {
		t.Fatalf("read: %v, wants %v", r, wants)
	}

	if e := buf.Rewind(1); e == nil {
		t.Fatalf("Rewind(1) must error")
	}
}

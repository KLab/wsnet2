package game

import (
	"fmt"
	"sync"
)

// MsgBuf rewindable ring buffer.
// Read/Write/Rewind can be called from different goroutines.
type MsgBuf struct {
	buf  []Msg
	mu   sync.RWMutex
	rSeq int
	wSeq int
}

// NewMsgBuf creates new MsgBuf.
// size: length of buffer.
func NewMsgBuf(size int) *MsgBuf {
	return &MsgBuf{
		buf: make([]Msg, size),
	}
}

func (b *MsgBuf) getSeqNo() (int, int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.rSeq, b.wSeq
}

// Write to buffer.
// It returns error when buffer is full.
func (b *MsgBuf) Write(data Msg) error {
	r, w := b.getSeqNo()
	s := len(b.buf)

	if w-s == r {
		return fmt.Errorf("MsgBuf overflow: size=%v, read=%v, write=%v", s, r, w)
	}

	b.buf[w%s] = data

	b.mu.Lock()
	b.wSeq++
	b.mu.Unlock()

	return nil
}

// Read returns all message stored in this buffer.
func (b *MsgBuf) Read() []Msg {
	r, w := b.getSeqNo()
	s := len(b.buf)
	count := w - r
	buf := make([]Msg, count)
	for i := 0; i < count; i++ {
		buf[i] = b.buf[(r+i)%s]
	}

	if count > 0 {
		b.mu.Lock()
		b.rSeq = w
		b.mu.Unlock()
	}

	return buf
}

// Rewind read sequence number.
func (b *MsgBuf) Rewind(seq int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	size := len(b.buf)
	if b.wSeq-seq >= size {
		return fmt.Errorf("MsgBuf too old seq num: %v, size:%v write:%v", seq, size, b.wSeq)
	}

	b.rSeq = seq
	return nil
}

package game

import (
	"fmt"
	"sync"
)

// EvBuf rewindable ring buffer.
// Read/Write/Rewind can be called from different goroutines.
type EvBuf struct {
	buf  []Event
	mu   sync.RWMutex
	rSeq int
	wSeq int
}

// NewEventBuf creates a new EvBuf.
// size: length of buffer.
func NewEvBuf(size int) *EvBuf {
	return &EvBuf{
		buf: make([]Event, size),
	}
}

func (b *EvBuf) getSeqNo() (int, int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.rSeq, b.wSeq
}

// Write to buffer from Room.MsgLoop goroutine.
// It returns an error when buffer is full.
func (b *EvBuf) Write(data Event) error {
	// MsgLoopは単一goroutineなのでロックし続ける必要はない
	r, w := b.getSeqNo()
	s := len(b.buf)

	if w-s == r {
		return fmt.Errorf("EvBuf overflow: size=%v, read=%v, write=%v", s, r, w)
	}

	b.buf[w%s] = data

	b.mu.Lock()
	b.wSeq++
	b.mu.Unlock()

	return nil
}

// Read returns all message stored in this buffer and last seqence numer.
// It called from Peer.EventLoop goroutine.
func (b *EvBuf) Read() ([]Event, int) {
	r, w := b.getSeqNo()
	s := len(b.buf)
	count := w - r
	buf := make([]Event, count)
	for i := 0; i < count; i++ {
		buf[i] = b.buf[(r+i)%s]
	}

	if count > 0 {
		b.mu.Lock()
		b.rSeq = w
		b.mu.Unlock()
	}

	return buf, w
}

// Rewind read sequence number.
func (b *EvBuf) Rewind(seq int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	size := len(b.buf)
	if b.wSeq-seq >= size {
		return fmt.Errorf("EvBuf too old seq num: %v, size:%v write:%v", seq, size, b.wSeq)
	}

	b.rSeq = seq
	return nil
}

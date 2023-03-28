package common

import (
	"sync"

	"golang.org/x/xerrors"
)

// RingBuf rewindable ring buffer.
// Read/Write can be called from different goroutines.
type RingBuf[T any] struct {
	buf  []T
	mu   sync.RWMutex
	rSeq int
	wSeq int

	hasData chan struct{}
}

// NewRingBuf creates a new RingBuf.
// size: length of buffer.
func NewRingBuf[T any](size int) *RingBuf[T] {
	return &RingBuf[T]{
		buf:     make([]T, size),
		hasData: make(chan struct{}, 1),
	}
}

// Write to buffer from Room.MsgLoop goroutine.
// It returns an error when buffer is full.
func (b *RingBuf[T]) Write(data T) error {
	// MsgLoopは単一goroutineで、wSeqはここでしか書き換えない
	// rSeqがwSeqを超えることは無いのでロックし続けなくてよい
	b.mu.RLock()
	r, w := b.rSeq, b.wSeq
	b.mu.RUnlock()

	s := len(b.buf)

	if w-s == r {
		return xerrors.Errorf("RingBuf overflow: size=%v, read=%v, write=%v", s, r, w)
	}

	b.buf[w%s] = data

	b.mu.Lock()
	b.wSeq++
	b.mu.Unlock()

	select {
	case b.hasData <- struct{}{}:
	default:
	}

	return nil
}

func (b *RingBuf[T]) HasData() <-chan struct{} {
	return b.hasData
}

// Read returns all message stored in this buffer and last seqence numer.
// It called from Client.EventLoop goroutine.
func (b *RingBuf[T]) Read(seq int) ([]T, error) {
	size := len(b.buf)

	b.mu.Lock()
	r, w := b.rSeq, b.wSeq
	if seq < r {
		// rewind read seq num
		if w-seq >= size {
			b.mu.Unlock()
			return nil, xerrors.Errorf("RingBuf too old seq num: %v, size:%v write:%v", seq, size, w)
		}
		b.rSeq = seq
		r = seq
	}
	b.mu.Unlock() // wSeqがrSeqを超えることは無いのでロックし続けなくて良い

	if r == w {
		return []T{}, nil
	}
	count := w - r
	buf := make([]T, count)
	for i := 0; i < count; i++ {
		buf[i] = b.buf[(r+i)%size]
	}

	b.mu.Lock()
	b.rSeq = w
	b.mu.Unlock()

	return buf, nil
}

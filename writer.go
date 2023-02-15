package bufwriter

import (
	"io"
	"sync"
	"time"
)

type Writer struct {
	len    int
	cap    int
	buf    []byte
	mu     sync.Mutex
	writer io.Writer
	done   chan bool
}

// New returns a new buffered writer
func New(capacity int, out io.Writer) *Writer {
	writer := Writer{
		buf:    make([]byte, capacity),
		cap:    capacity,
		mu:     sync.Mutex{},
		writer: out,
		done:   make(chan bool, 1),
	}

	return &writer
}

// StartFlusher starts the periodic buffer flushing goroutine
// TODO: swap callback to err chan
func (w *Writer) StartFlusher(interval time.Duration, errFn func(error)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.Flush(); err != nil && errFn != nil {
				errFn(err)
			}
		case <-w.done:
			w.done <- true
			return
		}
	}
}

// Stop stops the periodic buffer flushing (if enabled)
func (w *Writer) Stop() {
	if w.done != nil {
		w.done <- true
		<-w.done
		close(w.done)
	}
}

// Bytes returns the buffer byte slice
func (w *Writer) Bytes() []byte {
	w.mu.Lock()
	b := w.buf
	w.mu.Unlock()
	return b
}

// Write performs a buffered write to the destination io.Writer
func (w *Writer) Write(data []byte) (int, error) {
	sz := len(data)
	if sz == 0 {
		return 0, nil
	}

	n := w.Length()

	// Write buf + data if expecting buffer overflow
	if n+sz >= w.cap {
		if err := w.Flush(); err != nil {
			return 0, err
		}
		return w.writer.Write(data)
	}

	w.mu.Lock()
	w.len += copy(w.buf[w.len:], data)
	w.mu.Unlock()

	return sz, nil
}

// Flush writes the contents of the buffer into the destination io.Writer
func (w *Writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.len == 0 {
		return nil
	}

	n, err := w.writer.Write(w.buf[0:w.len])
	if err != nil {
		return err
	}
	if n != w.len {
		return io.ErrShortWrite
	}
	w.len = 0

	return nil
}

// Length returns the length of memory buffer
func (w *Writer) Length() int {
	w.mu.Lock()
	l := w.len
	w.mu.Unlock()
	return l
}

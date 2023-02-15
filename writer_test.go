package bufwriter

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	out := &dummyWriter{
		buf: bytes.NewBuffer(make([]byte, 1024)),
	}

	w := New(10, out)
	assert.Equal(t, 0, w.Length())

	n, err := w.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 0, out.Written())
	assert.Equal(t, 0, out.Calls())

	n, err = w.Write([]byte("world"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 10, out.Written())
	assert.Equal(t, 2, out.Calls())
}

func TestWriterWithFlushing(t *testing.T) {
	out := &dummyWriter{}
	w := New(1024, out)
	ts := time.Now()

	go w.StartFlusher(time.Millisecond*250, func(err error) { panic(err) })
	defer w.Stop()

	assert.Equal(t, 0, w.Length())
	assert.Equal(t, 0, out.Written())
	assert.Equal(t, 0, out.Calls())

	for i := 0; i < 10; i++ {
		w.Write([]byte(fmt.Sprintf("iter%d", i)))
	}

	assert.Less(t, time.Since(ts), time.Millisecond*250)
	assert.Equal(t, 50, w.Length())
	assert.Equal(t, 0, out.Written())
	assert.Equal(t, 0, out.Calls())

	time.Sleep(time.Millisecond * 260)
	assert.Greater(t, time.Since(ts), time.Millisecond*250)
	assert.Equal(t, 0, w.Length())
	assert.Equal(t, 50, out.Written())
	assert.Equal(t, 1, out.calls)
}

func BenchmarkWriter(b *testing.B) {
	out := &dummyWriter{}
	writer := New(1024*1024, out)
	data := make([]byte, 1000)

	for n := 0; n < b.N; n++ {
		_, err := writer.Write(data)
		if err != nil {
			panic(err)
		}
	}
}

type dummyWriter struct {
	mu    sync.Mutex
	calls int
	bytes int
	buf   *bytes.Buffer
}

func (w *dummyWriter) Calls() int {
	w.mu.Lock()
	n := w.calls
	w.mu.Unlock()
	return n
}

func (w *dummyWriter) Written() int {
	w.mu.Lock()
	n := w.bytes
	w.mu.Unlock()
	return n
}

func (w *dummyWriter) Bytes() []byte {
	w.mu.Lock()
	var b []byte
	if w.buf != nil {
		b = w.buf.Bytes()
	}
	w.mu.Unlock()
	return b
}

func (w *dummyWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.calls++
	w.bytes += len(data)

	if w.buf == nil {
		return len(data), nil
	}

	return w.buf.Write(data)
}

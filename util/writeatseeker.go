package util

import (
	"io"
	"sync"
)

type WriteAtSeeker interface {
	io.Writer
	io.WriterAt
	io.Seeker
}

type writeAtSeeker struct {
	w   io.WriteSeeker
	mux sync.Mutex
}

func NewWriteAtSeeker(w io.WriteSeeker) WriteAtSeeker {
	if was, ok := w.(WriteAtSeeker); ok {
		return was
	}

	return &writeAtSeeker{
		w: w,
	}
}

func (w *writeAtSeeker) Write(buf []byte) (int, error) {
	w.mux.Lock()
	defer w.mux.Unlock()
	return w.w.Write(buf)
}

func (w *writeAtSeeker) WriteAt(buf []byte, off int64) (int, error) {
	w.mux.Lock()
	defer w.mux.Unlock()

	// Switch positions
	returnTo, err := w.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, nil
	}
	defer w.w.Seek(returnTo, io.SeekStart)
	if _, err = w.w.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}

	// Write
	return w.w.Write(buf)
}

func (w *writeAtSeeker) Seek(offset int64, whence int) (int64, error) {
	w.mux.Lock()
	defer w.mux.Unlock()

	return w.w.Seek(offset, whence)
}

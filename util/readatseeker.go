package util

import (
	"io"
	"sync"
)

type ReadAtSeeker interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type readAtSeeker struct {
	r   io.ReadSeeker
	mux sync.Mutex
}

func NewReadAtSeeker(r io.ReadSeeker) ReadAtSeeker {
	if ras, ok := r.(ReadAtSeeker); ok {
		return ras
	}

	return &readAtSeeker{
		r: r,
	}
}

func (r *readAtSeeker) Read(buf []byte) (int, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	return io.ReadFull(r.r, buf)
}

func (r *readAtSeeker) ReadAt(buf []byte, off int64) (int, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	// Switch positions
	returnTo, err := r.r.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, nil
	}
	defer r.r.Seek(returnTo, io.SeekStart)
	if _, err = r.r.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}

	// Write
	return r.r.Read(buf)
}

func (r *readAtSeeker) Seek(offset int64, whence int) (int64, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	return r.r.Seek(offset, whence)
}

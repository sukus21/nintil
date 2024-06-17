package util

import (
	"fmt"
	"io"
)

// GO has no ready-made in-memory implementation io.WriteSeeker, so here's my version.
type WriteSeeker struct {
	Buf []byte
	Pos int64
}

func NewWriteSeeker(buf []byte) *WriteSeeker {
	return &WriteSeeker{
		Buf: buf,
		Pos: 0,
	}
}

func (w *WriteSeeker) Write(data []byte) (int, error) {
	n := copy(w.Buf[w.Pos:], data)
	if n != len(data) {
		return n, fmt.Errorf("trying to write outside buffer")
	}
	w.Pos += int64(len(data))
	return n, nil
}

func (w *WriteSeeker) Seek(offset int64, whence int) (int64, error) {
	npos := int64(0)
	switch whence {
	case io.SeekCurrent:
		npos = w.Pos + offset
	case io.SeekStart:
		npos = offset
	case io.SeekEnd:
		npos = int64(len(w.Buf)) + offset
	}

	if npos > int64(len(w.Buf)) || npos < 0 {
		return w.Pos, fmt.Errorf("trying to seek outside buffer")
	}
	w.Pos = npos
	return npos, nil
}

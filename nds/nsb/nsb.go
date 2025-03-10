package nsb

import (
	"encoding/binary"
	"io"
)

type NsbReader struct {
	io.ReadSeeker
	b binary.ByteOrder
}

func (r *NsbReader) GetEndian() binary.ByteOrder {
	return r.b
}

func newReader(r io.ReadSeeker) *NsbReader {
	return &NsbReader{
		ReadSeeker: r,
		b:          binary.LittleEndian,
	}
}

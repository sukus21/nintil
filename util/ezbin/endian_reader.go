package ezbin

import (
	"encoding/binary"
	"io"
)

var DefaultEndian binary.ByteOrder = binary.LittleEndian

type Endianed interface {
	GetEndian() binary.ByteOrder
}

type EndianedReader struct {
	io.Reader
	ByteOrder binary.ByteOrder
}

func (r *EndianedReader) GetEndian() binary.ByteOrder {
	return r.ByteOrder
}

type EndianedWriter struct {
	io.Writer
	ByteOrder binary.ByteOrder
}

func (w *EndianedWriter) GetEndian() binary.ByteOrder {
	return w.ByteOrder
}

// Get endian from reader/writer, if it implements the interface.
// Otherwise, return default endian.
func getEndian(v any) binary.ByteOrder {
	if endr, ok := v.(Endianed); ok {
		return endr.GetEndian()
	} else {
		return DefaultEndian
	}
}

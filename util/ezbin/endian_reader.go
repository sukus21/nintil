package ezbin

import (
	"encoding/binary"
	"io"
	"slices"
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

// Reads bytes as if an integer was read.
// If big endian, this is the same as reading normally.
// If little endian, the bytes will be in reversed before being returned.
func ReadWithOrder(r io.Reader, buf []byte, byteOrder binary.ByteOrder) (n int, err error) {
	n, err = io.ReadFull(r, buf)
	if err == nil && byteOrder == binary.LittleEndian {
		slices.Reverse(buf)
	}

	return n, err
}

// Reads bytes as if an integer was read.
// If big endian, this is the same as reading normally.
// If little endian, the bytes will be in reversed before being returned.
func (r *EndianedReader) ReadWithOrder(buf []byte) (n int, err error) {
	return ReadWithOrder(r, buf, r.ByteOrder)
}

type EndianedWriter struct {
	io.Writer
	ByteOrder binary.ByteOrder
}

func (w *EndianedWriter) GetEndian() binary.ByteOrder {
	return w.ByteOrder
}

// Writes bytes as if an integer was written.
// If big endian, this is the same as writing normally.
// If little endian, the bytes will be written in reverse order.
func (r *EndianedWriter) WriteWithOrder(buf []byte) (n int, err error) {
	if r.ByteOrder == binary.LittleEndian {
		newBuf := slices.Clone(buf)
		slices.Reverse(newBuf)
		buf = newBuf
	}

	return r.Write(buf)
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

func ReadBytesAsInt(r io.Reader, numBytes int) (uint64, error) {
	buf := [8]byte{}

	rawData := uint64(0)
	rawBytes := buf[:numBytes]
	_, err := ReadWithOrder(r, rawBytes, getEndian(r))
	if err != nil {
		return 0, err
	}
	for i := range rawBytes {
		rawData |= uint64(rawBytes[i]) << ((len(rawBytes) - (i + 1)) * 8)
	}

	return uint64(rawData), nil
}

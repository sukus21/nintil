package ezbin

import (
	"encoding/binary"
	"io"
)

// Read data from r into data.
// All entries in data should be pointers/references.
// Errors are ignored.
func Read(r io.Reader, data ...any) error {
	for _, n := range data {
		err := binary.Read(r, DefaultEndian, n)
		if err != nil {
			return err
		}
	}
	return nil
}

// Same as Read, but reads AT in a io.ReaderAt.
// Errors are ignored.
func ReadAt[K Integer](r io.ReaderAt, at K, data ...any) error {
	sect := EndianedReader{
		Reader:    io.NewSectionReader(r, int64(at), -1),
		ByteOrder: getEndian(r),
	}
	return Read(sect, data...)
}

// Write all data into w.
// Errors are ignored.
func Write(w io.Writer, data ...any) error {
	for _, n := range data {
		err := binary.Write(w, DefaultEndian, n)
		if err != nil {
			return err
		}
	}
	return nil
}

// Same as Write, but with WriterAt.
// Errors are ignored.
func WriteAt[K Integer](w io.WriterAt, at K, data ...any) error {
	ofsw := EndianedWriter{
		Writer:    io.NewOffsetWriter(w, int64(at)),
		ByteOrder: getEndian(w),
	}

	return Write(ofsw, data...)
}

// Return single read value.
// Errors are ignored.
func ReadSingle[K any](r io.Reader) K {
	var d K
	binary.Read(r, getEndian(r), &d)
	return d
}

type Integer interface {
	~int | ~uint | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64 | ~uintptr
}

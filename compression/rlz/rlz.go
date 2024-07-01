package rlz

import (
	"bytes"
	"io"
)

// Reads a compressed stream of RLZ compressed data.
// To learn more about the format, check the README file.
type RLZReader interface {
	io.Reader
	io.ByteReader

	// Returns the total size of the decompressed data.
	Size() int
}

// Decompresses a blob of RLZ compressed data.
func Decompress(dat []byte) ([]byte, error) {
	buf := bytes.NewReader(dat)
	dec, err := NewReader(buf)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(dec)
}

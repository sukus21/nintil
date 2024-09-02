package rlz

import (
	"bytes"
	"errors"
	"io"
)

var ErrRlzSize = errors.New("rlz: decompressed data size differs from size specified in header")

// Reads a compressed stream of RLZ compressed data.
// To learn more about the format, check the README file.
type RLZReader interface {
	io.Reader
	io.ByteReader

	// Returns the total size of the decompressed data.
	Size() int
}

// Decompresses a blob of RLZ compressed data.
// As a bonus, this method validates the size of the output data.
func Decompress(dat []byte) ([]byte, error) {
	buf := bytes.NewReader(dat)
	dec, err := NewReader(buf)
	if err != nil {
		return nil, err
	}

	// Decompress all data
	out, err := io.ReadAll(dec)
	if err != nil {
		return nil, err
	}

	// Validate size
	if dec.Size() != len(out) {
		return out, ErrRlzSize
	}
	return out, nil
}

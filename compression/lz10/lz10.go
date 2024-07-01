package lz10

import (
	"bytes"
	"errors"
	"io"
)

var ErrNotLZ10 = errors.New("data is not LZ10 compressed")

// Decompresses a blob of LZ10 compressed data.
// If data is not LZ10 compressed, ErrNotLZ10 is returned.
func Decompress(dat []byte) ([]byte, error) {
	buf := bytes.NewReader(dat)
	dec, err := NewReader(buf)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(dec)
}

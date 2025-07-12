package pmoc

import (
	"bytes"
	"io"

	"github.com/sukus21/nintil/util"
)

// Creates a reader from a PMOC compressed reader.
// The source reader passed in is now owned by RLXReader.
//
// This format is tricky to make a proper streamed implementation for.
// Instead, decompress it on the spot, then return a bytes.Reader...
//
// A proper streamed solution might be implemented some day.
// When that day comes, the interface to this function will not change.
func NewReader(src io.Reader, sizeCompressed uint32) (io.Reader, error) {
	buf := make([]byte, sizeCompressed)
	_, err := io.ReadFull(src, buf)
	if err != nil {
		return nil, util.TranslateEOF(err)
	}

	decompressed, err := Decompress(buf)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(decompressed), nil
}

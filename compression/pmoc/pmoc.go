package pmoc

import (
	"bytes"
	"errors"
	"io"

	"github.com/sukus21/nintil/compression/rlx"
	"github.com/sukus21/nintil/util/ezbin"
)

const magicPMOC = 0x434F4D50

var ErrNotPMOC = errors.New("data is not PMOC compressed")

// Decompresses an PMOC compressed blob.
//
// This code is lifted straight from NitroPaint.
// https://github.com/Garhoogin/NitroPaint/blob/93a460b85f71fc46fa53f763f41ca7dd29c68699/NitroPaint/compression.c#L1557
func Decompress(src []byte) ([]byte, error) {
	r := bytes.NewReader(src)

	magic := ezbin.ReadSingle[uint32](r)
	if magic != magicPMOC || len(src) < 16 {
		return nil, ErrNotPMOC
	}

	totalSize := ezbin.ReadSingle[uint32](r)
	numSegments := ezbin.ReadSingle[uint32](r)
	compressedLength := ezbin.ReadSingle[uint32](r)
	_ = compressedLength

	segmentLengths := make([]uint32, numSegments)
	if err := ezbin.Read(r, segmentLengths); err != nil {
		return nil, err
	}

	w := make([]byte, totalSize)
	i := 0
	for _, segmentLength := range segmentLengths {

		// If length is < 0, this segment is not compressed
		if segmentLength <= 0 {
			io.ReadFull(r, w[i:i+int(segmentLength)])
			i += int(segmentLength)
			continue
		}

		// This is an RLX compressed segment. Decompress and append
		buf := make([]byte, segmentLength)
		io.ReadFull(r, buf)
		out, err := rlx.Decompress(buf)
		if err != nil {
			return nil, err
		}

		copy(w[i:], out)
		i += len(out)
	}

	return w, nil
}

package pit

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/sukus21/nintil/util/ezbin"
)

// Open a dat file from a reader
func OpenDat(r io.ReaderAt, index int) (*io.SectionReader, error) {
	var buf [8]byte
	if _, err := r.ReadAt(buf[:], int64(index)*4); err != nil {
		return nil, err
	}
	return io.NewSectionReader(
		r,
		int64(binary.LittleEndian.Uint32(buf[:])),
		int64(binary.LittleEndian.Uint32(buf[4:])),
	), nil
}

// Get number of dats inside a single file
func NumDats(r io.ReaderAt, length int) int {
	var offset, previous uint32
	for i := 0; ; i += 4 {
		if err := ezbin.ReadAt(r, i, &offset); err != nil {
			return 0
		}

		// Not a .dat file
		if offset < previous {
			return 0
		}

		if offset == uint32(length) {
			return i << 2
		}
	}
}

// Check if a binary blob is a .dat file
func IsDat(dat []byte) bool {
	first := 0
	prev := 0
	for i := 0; i+3 < len(dat); i += 4 {
		offset := int(binary.LittleEndian.Uint32(dat[i : i+4]))

		// Full file contents used, is a .dat file
		if i >= 4 && int(offset) == len(dat) {
			return i+4 == first
		}

		// Entry beyond file length, not a .dat file
		if int(offset) > len(dat) {
			return false
		}

		// Entry did not come in sequence, not a .dat
		if offset < prev && prev != 0 {
			return false
		}
		prev = offset
		if i == 0 {
			first = offset
		}
	}

	// Abrupt end, not a .dat file
	return false
}

// Unpack contents of .dat files
func UnpackDat(dat []byte) [][]byte {

	// Get list of offsets
	offsets := make([]uint32, 0)
	l := len(dat)
	for i := 0; ; i += 4 {
		o := binary.LittleEndian.Uint32(dat[i : i+4])
		offsets = append(offsets, o)
		if o == uint32(l) {
			break
		}
	}

	// Read all file contents
	files := make([][]byte, 0)
	for i := 0; i < len(offsets)-1; i++ {
		d := dat[offsets[i]:offsets[i+1]]
		files = append(files, d)
	}

	// Return that
	return files
}

// Unpack a single file from a .dat file
func UnpackDatSingle(dat []byte, i int) []byte {
	offset := binary.LittleEndian.Uint32(dat[i*4:])
	offsetLength := binary.LittleEndian.Uint32(dat[i*4+4:])
	return dat[offset:offsetLength]
}

func ExportDat(b []byte, path string) {
	os.MkdirAll(path, os.ModePerm)
	f := UnpackDat(b)
	for i, v := range f {
		os.WriteFile(fmt.Sprintf("%s/%d", path, i), v, os.ModePerm)
	}
}

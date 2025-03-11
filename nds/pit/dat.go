package pit

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

var ErrNotDat = errors.New("invalid .dat archive")

// Open a dat file from a reader
func OpenDat(r io.ReaderAt, index int) (*io.SectionReader, error) {
	var buf [8]byte
	if _, err := r.ReadAt(buf[:], int64(index)*4); err != nil {
		return nil, err
	}

	start := binary.LittleEndian.Uint32(buf[:])
	end := binary.LittleEndian.Uint32(buf[4:])
	return io.NewSectionReader(
		r,
		int64(start),
		int64(end-start),
	), nil
}

// Read the file offsets in a .dat file
func DatOffsets(dat []byte) []uint32 {
	offsets := make([]uint32, 0, 32)
	first := uint32(0)
	prev := uint32(0)
	for i := uint32(0); i+3 < uint32(len(dat)); i += 4 {
		offset := binary.LittleEndian.Uint32(dat[i : i+4])
		offsets = append(offsets, offset)

		// Full file contents used, is a .dat file
		if i >= 4 && int(offset) == len(dat) {
			if i+4 == first {
				return offsets
			}
		}

		// Entry beyond file length, not a .dat file
		if int(offset) > len(dat) {
			return nil
		}

		// Entry did not come in sequence, not a .dat
		if offset < prev && prev != 0 {
			return nil
		}
		prev = offset
		if i == 0 {
			first = offset
		}
	}

	// Abrupt end, not a .dat file
	return nil
}

// Get number of dats inside a single file
func NumDats(dat []byte) int {
	return max(0, len(DatOffsets(dat))/4-1)
}

// Check if a binary blob is a .dat file
func IsDat(dat []byte) bool {
	return NumDats(dat) != 0
}

// Unpack contents of .dat files
func UnpackDat(dat []byte) [][]byte {

	// Get list of offsets
	offsets := DatOffsets(dat)

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

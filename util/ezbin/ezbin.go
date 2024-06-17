package ezbin

import (
	"bytes"
	"encoding/binary"
	"io"
)

// Read data from r as little endian into data.
// All entries in data should be pointers/references.
func Read(r io.Reader, data ...any) error {
	for _, n := range data {
		err := binary.Read(r, binary.LittleEndian, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadAt[K Integer](r io.ReaderAt, at K, data ...any) error {
	sect := io.NewSectionReader(r, int64(at), -1)
	return Read(sect, data...)
}

// Read data from r as big endian into data.
// All entries in data should be pointers/references.
func ReadBigEnd(r io.Reader, data ...any) error {
	for _, n := range data {
		err := binary.Read(r, binary.BigEndian, n)
		if err != nil {
			return err
		}
	}
	return nil
}

// Write all data into w as little endian.
func Write(w io.Writer, data ...any) error {
	for _, n := range data {
		err := binary.Write(w, binary.LittleEndian, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteAt[K Integer](w io.WriterAt, at K, data ...any) error {
	ofsw := io.NewOffsetWriter(w, int64(at))
	return Write(ofsw, data...)
}

// Write all data into w as little endian.
func WriteBigEnd(w io.Writer, data ...any) error {
	for _, n := range data {
		err := binary.Write(w, binary.BigEndian, n)
		if err != nil {
			return err
		}
	}
	return nil
}

// Always uses little endian.
func ReadSingle[K any](r io.Reader) K {
	var d K
	binary.Read(r, binary.LittleEndian, &d)
	return d
}

// Always uses big endian.
func ReadSingleBigEnd[K any](r io.Reader) K {
	var d K
	binary.Read(r, binary.BigEndian, &d)
	return d
}

// Create a pre-initialized array.
func FillerArray[K any](length int, value K) []K {
	buf := make([]K, length)
	for i := range buf {
		buf[i] = value
	}
	return buf
}

func Put(buf []byte, pos int, data ...any) ([]byte, error) {

	// Write data to temp buffer
	tmp := &bytes.Buffer{}
	err := Write(tmp, data...)
	if err != nil {
		return nil, err
	}

	// Expand existing buffer if needed
	tempd := tmp.Bytes()
	if pos+len(tempd) > len(buf) {
		buf = append(buf, make([]byte, pos+len(tempd)-len(buf))...)
	}

	// Copy buffered data into real buffer
	copy(buf[pos:], tempd)
	return buf, nil
}

func PutAny[K any](buf []K, pos int, data ...K) []K {
	if len(buf) < pos+len(data) {
		buf = append(buf, make([]K, (pos+len(data))-len(buf))...)
	}

	// Copy buffered data into real buffer
	copy(buf[pos:], data)
	return buf
}

func Get(buf []byte, pos int, data ...any) error {
	return Read(bytes.NewReader(buf[pos:]), data...)
}

type Integer interface {
	~int | ~uint | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64 | ~uintptr
}

// Returns the amount of padding needed to align number
func Pad[K Integer](n K, lim K) K {
	return (lim - (n & (lim - 1))) & (lim - 1)
}

// Returns the number padded
func PadTo[K Integer](n K, lim K) K {
	return n + Pad(n, lim)
}

// Assumes the write head is aligned.
// If you want to make sure it is, use the Align function.
func WritePadded(w io.Writer, lim int, filler byte, data ...any) error {
	b := &bytes.Buffer{}
	if err := Write(b, data...); err != nil {
		return err
	}
	dat := b.Bytes()
	padLength := Pad(len(dat), lim)
	dat = append(dat, FillerArray(padLength, filler)...)
	_, err := w.Write(dat)
	return err
}

// Get current position of seeker, but as any integer type
func At[K Integer](w io.Seeker) (K, error) {
	return Seek(w, K(0), io.SeekCurrent)
}

// Aligns seeker head to match lim (rounds UP)
func Align[K Integer](s io.Seeker, lim K) (K, error) {
	pos, _ := s.Seek(0, io.SeekCurrent)
	p := Pad(pos, int64(lim))
	return Seek(s, K(pos+p), io.SeekStart)
}

// io.Seeker.Seek() but with generic offset types
func Seek[K Integer](s io.Seeker, offset K, whence int) (K, error) {
	pos, err := s.Seek(int64(offset), whence)
	return K(pos), err
}

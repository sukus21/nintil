package ezbin

import (
	"bytes"
	"io"
)

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

package lz10

import (
	"encoding/binary"
	"io"

	"github.com/sukus21/nintil/util/ezbin"
)

type LZ10Reader struct {
	rewind    [0x1000]byte
	rewindPos int
	buf       []byte
	src       io.Reader
	blocks    uint32
}

// Creates a reader from an LZ10 compressed reader.
// The source reader passed in is now owned by LZ10Reader.
// If data is not LZ10 compressed, returns ErrNotLZ10.
func NewReader(src io.Reader) (io.Reader, error) {
	h := uint32(0)
	binary.Read(src, binary.LittleEndian, &h)
	magic := h & 0xFF
	if magic != 0x10 {
		return nil, ErrNotLZ10
	}

	return &LZ10Reader{
		src:    src,
		blocks: h >> 8,
	}, nil
}

func (r *LZ10Reader) readSrcByte() (byte, error) {
	instb := [1]byte{}
	if _, err := io.ReadFull(r.src, instb[:]); err != nil {
		return 0, err
	}
	return instb[0], nil
}

func (r *LZ10Reader) readBuffered(buf []byte) (int, error) {
	n := copy(buf, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

func (r *LZ10Reader) getRewind(offset int) byte {
	pos := (r.rewindPos - offset) & 0x0FFF
	return r.rewind[pos]
}

func (r *LZ10Reader) appendData(dat byte) {
	r.buf = append(r.buf, dat)
	pos := r.rewindPos & 0x0FFF
	r.rewind[pos] = dat
	r.rewindPos++
}

func (r *LZ10Reader) Read(buf []byte) (int, error) {
	// Empty buffered content
	if len(r.buf) != 0 {
		return r.readBuffered(buf)
	}

	// No more blocks?
	if r.blocks == 0 {
		return 0, io.EOF
	}

	inst, err := r.readSrcByte()
	if err != nil {
		return 0, err
	}

	// Decompress data
	r.buf = make([]byte, 0, 0x1000)
	for i := 0; i < 8; i++ {
		if inst&0x80 != 0 {
			conf := ezbin.ReadSingleBigEnd[uint16](r.src)

			// Rewind and copy old data
			rewindCount := int((conf >> 12) + 3)
			offset := int(conf & 0x0FFF)
			for i := 0; i < rewindCount; i++ {
				r.appendData(r.getRewind(offset + 1))
			}
			r.blocks -= uint32(rewindCount)
		} else {

			// New single byte
			v, err := r.readSrcByte()
			if err != nil {
				return 0, err
			}
			r.appendData(v)
			r.blocks--
		}

		inst <<= 1
		if r.blocks == 0 {
			break
		}
	}

	// Read data written to buffer
	return r.readBuffered(buf)
}

func (r *LZ10Reader) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := r.Read(buf[:])
	return buf[0], err
}

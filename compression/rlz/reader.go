package rlz

import (
	"encoding/binary"
	"io"
)

type rlzReader struct {
	rewind    [0x1000]byte
	rewindPos int
	buf       []byte
	src       io.Reader
	blocks    uint32
	size      uint32
}

func NewReader(src io.Reader) (RLZReader, error) {
	r := &rlzReader{
		src: src,
	}

	// Get decompressed file size
	h, err := r.readSrcByte()
	if err != nil {
		return nil, err
	}
	r.size = uint32(h & 0x3F)
	varlen := h >> 6
	for i := 6; varlen > 0; i += 6 {
		n, err := r.readSrcByte()
		if err != nil {
			return nil, err
		}
		r.size |= uint32(n) << i
		varlen--
	}

	// Get block count
	h, err = r.readSrcByte()
	if err != nil {
		return nil, err
	}
	r.blocks = uint32(h&0x3F) + 1
	varlen = h >> 6
	for i := 6; varlen > 0; i += 6 {
		n, err := r.readSrcByte()
		if err != nil {
			return nil, err
		}
		r.blocks |= uint32(n) << i
		varlen--
	}

	return r, nil
}

func (r *rlzReader) readSrcByte() (byte, error) {
	instb := [1]byte{}
	if _, err := r.src.Read(instb[:]); err != nil {
		return 0, err
	}
	return instb[0], nil
}

func (r *rlzReader) readBuffered(buf []byte) (int, error) {
	n := copy(buf, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

func (r *rlzReader) getRewind(offset int) byte {
	pos := (r.rewindPos - offset) & 0x0FFF
	return r.rewind[pos]
}

func (r *rlzReader) appendData(dat byte) {
	r.buf = append(r.buf, dat)
	pos := r.rewindPos & 0x0FFF
	r.rewind[pos] = dat
	r.rewindPos++
}

func (r *rlzReader) decompress() error {
	for {
		commandFrame, err := r.readSrcByte()
		if err != nil {
			return err
		}
		for i := 0; i < 8; i += 2 {
			command := (commandFrame >> i) & 0x03
			switch command {

			// End of block
			case 0x00:
				return nil

			// Single byte
			case 0x01:
				data, err := r.readSrcByte()
				if err != nil {
					return err
				}
				r.appendData(data)

			// Rewind encoding, LZ-like
			case 0x02:
				rewindBase, _ := r.readSrcByte()
				operand, err := r.readSrcByte()
				if err != nil {
					return err
				}

				rewindCount := 2 + uint32(operand&0x0F)
				rewindOffset := uint32(rewindBase) + uint32(operand&0xF0)<<4
				for i := uint32(0); i < rewindCount; i++ {
					r.appendData(r.getRewind(int(rewindOffset)))
				}

			// Run-length encoding
			case 0x03:
				count, _ := r.readSrcByte()
				data, err := r.readSrcByte()
				if err != nil {
					return err
				}
				for i := -2; i < int(count); i++ {
					r.appendData(data)
				}
			}
		}
	}
}

func (r *rlzReader) Read(buf []byte) (int, error) {
	if len(r.buf) != 0 {
		return r.readBuffered(buf)
	}

	// No more blocks?
	if r.blocks == 0 {
		return 0, io.EOF
	}

	// Allocate new buffer
	blockSize := uint16(0)
	err := binary.Read(r.src, binary.LittleEndian, &blockSize)
	if err != nil {
		return 0, err
	}
	r.buf = make([]byte, 0, blockSize)

	// Decompress data
	r.blocks--
	if err := r.decompress(); err != nil {
		return 0, err
	}

	// Now we have some buffered data, return that
	return r.readBuffered(buf)
}

func (r *rlzReader) ReadByte() (byte, error) {
	buf := [1]byte{}
	_, err := r.Read(buf[:])
	return buf[0], err
}

func (r *rlzReader) Size() int {
	return int(r.size)
}

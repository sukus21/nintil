package rlx

import (
	"encoding/binary"
	"errors"
	"io"
)

var ErrInvalidOffset = errors.New("rlx: repeat offset is out of bounds")

// Decompresses an RLX compressed blob.
//
// This code is lifted straight from NitroPaint.
// https://github.com/Garhoogin/NitroPaint/blob/93a460b85f71fc46fa53f763f41ca7dd29c68699/NitroPaint/compression.c#L652
func Decompress(buffer []byte) ([]byte, error) {
	// decompress the input buffer.
	if len(buffer) < 4 {
		return nil, io.ErrUnexpectedEOF
	}

	// find the length of the decompressed buffer.
	length := binary.LittleEndian.Uint32(buffer) >> 8

	// create a buffer for the decompressed buffer
	result := make([]byte, length)

	// initialize variables
	srcOffset := uint32(4)
	dstOffset := uint32(0)

	for {
		head := buffer[srcOffset]
		srcOffset++

		// loop 8 times
		for range 8 {
			flag := head >> 7
			head <<= 1

			// Fresh single byte
			if flag == 0 {
				result[dstOffset] = buffer[srcOffset]
				dstOffset++
				srcOffset++
				if dstOffset == length {
					return result, nil
				}
				continue
			}

			if int(srcOffset+2) > len(buffer) {
				return nil, io.ErrUnexpectedEOF
			}

			// Compressed data
			high := uint32(buffer[srcOffset])
			srcOffset++
			low := uint32(buffer[srcOffset])
			srcOffset++
			mode := high >> 4

			repeatLength := uint32(0)
			repeatOffset := uint32(0)
			switch mode {

			// 8-bit length +0x11, 12-bit offset
			case 0:
				if int(srcOffset+1) > len(buffer) {
					return nil, io.ErrUnexpectedEOF
				}
				low2 := uint32(buffer[srcOffset])
				srcOffset++
				repeatLength = ((high << 4) | (low >> 4)) + 0x11
				repeatOffset = (((low & 0xF) << 8) | low2) + 1

			// 16-bit length +0x111, 12-bit offset
			case 1:
				if int(srcOffset+2) > len(buffer) {
					return nil, io.ErrUnexpectedEOF
				}
				low2 := uint32(buffer[srcOffset])
				srcOffset++
				low3 := uint32(buffer[srcOffset])
				srcOffset++
				repeatLength = (((high & 0xF) << 12) | (low << 4) | (low2 >> 4)) + 0x111 //
				repeatOffset = (((low2 & 0xF) << 8) | low3) + 1

			// 4-bit length +0x1 (but >= 3), 12-bit offset
			default:
				repeatLength = (high >> 4) + 1
				repeatOffset = (((high & 0xF) << 8) | low) + 1
			}

			// Hand on... is this OK?
			if int(dstOffset)-int(repeatOffset) < 0 {
				return nil, ErrInvalidOffset
			}

			//write back
			for j := uint32(0); j < repeatLength; j++ {
				result[dstOffset] = result[dstOffset-repeatOffset]
				dstOffset++
				if dstOffset == length {
					return result, nil
				}
			}
		}
	}
}

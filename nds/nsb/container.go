package nsb

import "encoding/binary"

type Container struct {
	StartPos      int64            `ezbin_seekpos:""`
	Stamp         string           `ezbin_string:"ascii" ezbin_length:"4"`
	ByteOrder     binary.ByteOrder `ezbin_byteorder:""`
	Version       uint16
	Filesize      uint32
	HeaderSize    uint16
	NumSubfiles   uint16
	SubfileOffets []uint32 `ezbin_length:"NumSubfiles"`
}

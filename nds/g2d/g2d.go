package g2d

import (
	"encoding/binary"
)

type G2DFile struct {
	Stamp      string           `ezbin_string:"ascii,4"`
	ByteOrder  binary.ByteOrder `ezbin_byteorder:""`
	Version    uint16
	FileSize   uint32
	HeaderSize uint16
	Blocks     []struct {
		Offset     struct{} `ezbin_tell:"block"`
		Stamp      string   `ezbin_string:"ascii,4"`
		DataLength uint32
		Data       []byte `ezbin_length:"DataLength,block"`
	} `ezbin_length:"u16"`
}

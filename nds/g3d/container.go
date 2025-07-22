package g3d

import (
	"encoding/binary"
)

type Container struct {
	_          struct{}         `ezbin_tell:"container"`
	Stamp      string           `ezbin_string:"ascii,4"`
	ByteOrder  binary.ByteOrder `ezbin_byteorder:""`
	Version    uint16
	Filesize   uint32
	HeaderSize uint16
	Subfiles   []subfileWrapper `ezbin_length:"u16" ezbin_offset_array:"u32,container"`
}

package nsb

import "encoding/binary"

type Container struct {
	_          struct{}         `ezbin_tell:"container"`
	Stamp      string           `ezbin_string:"ascii,4"`
	ByteOrder  binary.ByteOrder `ezbin_byteorder:""`
	Version    uint16
	Filesize   uint32
	HeaderSize uint16
	Subfiles   []Subfile `ezbin_length:"u16" ezbin_offset_array:"u32,container"`
}

type Subfile struct {
	Offset int64  `ezbin_tell:""`
	Stamp  string `ezbin_string:"ascii,4"`
}

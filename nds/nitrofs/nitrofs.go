package nitrofs

import (
	"io"
	"io/fs"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/mapping"
)

// NitroFS mapping template strings
const (
	mappingFileUnnamed = "NitroFS file #%02d"
	mappingFileNamed   = "NitroFS file: %s"
	mappingFNT         = "file name table"
	mappingFAT         = "file allocation table"
	mappingOVT9        = "ARM9 overlay table"
	mappingOVT7        = "ARM7 overlay table"
)

func FromROM(r io.ReadSeeker, info *Info, mmap *mapping.Mapping) NitroFS {
	nfs := &streamFS{
		info: info,
		r:    util.NewReadAtSeeker(r),
	}

	if mmap != nil {
		nfs.updateMapping(mmap)
	}

	return nfs
}

type NitroFS interface {
	fs.FS
	GetArm9Overlays() []Overlay
	GetArm7Overlays() []Overlay
}

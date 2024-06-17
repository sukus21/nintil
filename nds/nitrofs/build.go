package nitrofs

import (
	"errors"
	"io/fs"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/mapping"
)

func Build(w util.WriteAtSeeker, fsys fs.FS, mmap *mapping.Mapping) (info *Info, err error) {
	return nil, errors.New("not implemented")
}

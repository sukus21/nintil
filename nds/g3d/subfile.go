package g3d

import (
	"io"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

type Subfile interface {
	GetStamp() string
}

type subfileHeader struct {
	Offset int64  `ezbin_tell:""`
	Stamp  string `ezbin_string:"ascii,4"`
}

func (f *subfileHeader) GetStamp() string {
	return f.Stamp
}

type subfileWrapper struct {
	Subfile
}

func (f *subfileWrapper) EzbinDecode(r io.Reader) (err error) {
	defer util.Recover(&err)

	rs, ok := r.(io.ReadSeeker)
	if !ok {
		return ezbin.ErrNotSeeker
	}

	offset := util.Must1(ezbin.At[int64](rs))
	stamp := util.Must1(ezbin.ReadString(r, ezbin.StringFormat_ASCII, 4))
	util.Must1(rs.Seek(offset, io.SeekStart))

	switch stamp {

	case "MDL0":
		f.Subfile = util.Must1(ezbin.Decode[*SubfileMDL0](rs))
	case "TEX0":
		f.Subfile = util.Must1(ezbin.Decode[*SubfileTEX0](rs))

	// What even is this?
	default:
		f.Subfile = &subfileHeader{offset, stamp}
	}

	return nil
}

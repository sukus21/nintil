package ezbin

import (
	"bytes"
	"io"

	"github.com/sukus21/nintil/util"
)

type StringFormat int

const (
	StringFormat_Invalid = StringFormat(0)
	StringFormat_ASCII
)

func StringFormatFromString(format string) StringFormat {
	switch format {
	case "ascii":
		return StringFormat_ASCII

	default:
		return StringFormat_Invalid
	}
}

func ReadString(r io.Reader, format StringFormat, fixedLength int) (string, error) {

	// Do we read string from a fixed-length buffer?
	stringReader := r
	if fixedLength != -1 {
		dat := make([]byte, fixedLength)
		util.Must1(io.ReadFull(r, dat))
		stringReader = bytes.NewReader(dat)
	}

	str := ""
	switch format {
	case StringFormat_ASCII:
		for {
			b := ReadSingle[byte](stringReader)
			if b != 0 {
				str += string(rune(b))
			} else {
				break
			}
		}

	// TODO: support more string types
	default:
		panic(ErrUnknownString)
	}

	return str, nil
}

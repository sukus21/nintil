package nsb

import (
	"slices"
)

type Namelist[T any] struct {
	_               uint8
	Count           uint8
	Size            uint16
	_               uint16
	_               uint16
	_               uint32
	_               []uint32 `ezbin_length:"Count"`
	ElementSize     uint16
	DataSectionSize uint16
	Data            []T        `ezbin_length:"Count"`
	Names           [][16]byte `ezbin_length:"Count"`
}

func (n *Namelist[T]) Map() map[string]T {
	out := make(map[string]T)
	for i := range n.Data {
		strLen := slices.Index(n.Names[i][:], 0)
		if strLen == -1 {
			strLen = 16
		}
		out[string(n.Names[i][:strLen])] = n.Data[i]
	}
	return out
}

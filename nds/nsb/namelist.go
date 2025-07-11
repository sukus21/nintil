package nsb

type NameList[T any] struct {
	_     uint8
	Count uint8
	Size  uint16
	_     uint16
	_     uint16
	_     uint32
	_     []uint32 `ezbin_length:"Count"`

	ElementSize     uint16
	DataSectionTime uint16

	Data  []T      `ezbin_length:"Count" ezbin_offset_array:"u32,namelist_offset"`
	Names []string `ezbin_length:"Count" ezbin_string:"ascii,16"`
}

func (n *NameList[T]) Map() map[string]T {
	out := make(map[string]T)
	for i := range n.Data {
		out[n.Names[i]] = n.Data[i]
	}
	return out
}

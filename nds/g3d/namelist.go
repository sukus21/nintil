package g3d

type NamedItem[T any] struct {
	Name string
	Item T
}

type NameListOffset[T any] struct {
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

func (n *NameListOffset[T]) Map() map[string]T {
	out := make(map[string]T)
	for i := range n.Data {
		out[n.Names[i]] = n.Data[i]
	}
	return out
}

func (n *NameListOffset[T]) Items() []NamedItem[T] {
	out := make([]NamedItem[T], len(n.Data))
	for i := range n.Data {
		out[i] = NamedItem[T]{
			Name: n.Names[i],
			Item: n.Data[i],
		}
	}
	return out
}

type NameListImmediate[T any] struct {
	_     uint8
	Count uint8
	Size  uint16
	_     uint16
	_     uint16
	_     uint32
	_     []uint32 `ezbin_length:"Count"`

	ElementSize     uint16
	DataSectionTime uint16

	Data  []T      `ezbin_length:"Count"`
	Names []string `ezbin_length:"Count" ezbin_string:"ascii,16"`
}

func (n *NameListImmediate[T]) Map() map[string]T {
	out := make(map[string]T)
	for i := range n.Data {
		out[n.Names[i]] = n.Data[i]
	}
	return out
}

func (n *NameListImmediate[T]) Items() []NamedItem[T] {
	out := make([]NamedItem[T], len(n.Data))
	for i := range n.Data {
		out[i] = NamedItem[T]{
			Name: n.Names[i],
			Item: n.Data[i],
		}
	}
	return out
}

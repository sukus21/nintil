package mapping

import (
	"fmt"
)

func NewMapping(maxLength uint32) *Mapping {
	return &Mapping{
		mappings:  make([]*MappingEntry, 0, 256),
		maxLength: maxLength,
	}
}

type Mapping struct {
	mappings  []*MappingEntry
	maxLength uint32
}

// Will return nil if length is 0
func (m *Mapping) Add(name string, length uint32) (*MappingEntry, error) {
	if length <= 0 {
		return nil, nil
	}
	entry := &MappingEntry{
		name:   name,
		length: length,
	}
	return entry, m.insert(entry)
}

// Will return nil if length is 0
func (m *Mapping) AddAt(name string, at, length uint32) (*MappingEntry, error) {
	if length <= 0 {
		return nil, nil
	}
	entry := &MappingEntry{
		name:   name,
		length: length,
	}
	return entry, m.insertAt(entry, at)
}

func (m *Mapping) insert(entry *MappingEntry) error {
	pos := uint32(0)
	for i, v := range m.mappings {
		if v.from+pos <= entry.length {
			m.placeInto(entry, uint32(i))
			entry.from = v.from + pos
			return nil
		}
		pos = v.from + v.length
	}

	// Still nothing found?
	if m.maxLength-pos <= entry.length {
		m.mappings = append(m.mappings, entry)
		entry.from = pos
		return nil
	}

	// No space in ROM
	return fmt.Errorf("no more space in ROM")
}

func (m *Mapping) insertAt(entry *MappingEntry, at uint32) error {
	for i, v := range m.mappings {
		if v.To() <= at {
			continue
		}
		if v.from >= at && v.To() < at {
			return fmt.Errorf("space already occupied at %08X by %s", at, v.name)
		}
		if v.from-at < entry.length {
			return fmt.Errorf("not enough space at %08X before %s", at, v.name)
		}

		m.placeInto(entry, uint32(i))
		entry.from = at
		return nil
	}

	// Still nothing found?
	if at+entry.length <= m.maxLength {
		m.mappings = append(m.mappings, entry)
		entry.from = at
		return nil
	}

	return fmt.Errorf("no more space in ROM")
}

func (m *Mapping) placeInto(entry *MappingEntry, pos uint32) {

	// Expand mapping buffer, this will be overwritten in a moment
	m.mappings = append(m.mappings, entry)

	// Scoot everything over
	copy(m.mappings[pos+1:], m.mappings[pos:])

	// Insert new entry
	m.mappings[pos] = entry
}

func (m *Mapping) Find(pos uint32) *MappingEntry {
	for _, v := range m.mappings {
		if v.from <= pos && v.To() > pos {
			return v
		}
	}
	return nil
}

func (m *Mapping) String() string {
	res := ""
	for _, v := range m.mappings {
		res += fmt.Sprintf("offset 0x%08X -> 0x%08X: %s\n",
			v.from,
			v.To(),
			v.name,
		)
	}
	return res
}

type MappingEntry struct {
	name   string
	from   uint32
	length uint32
}

func (m *MappingEntry) From() uint32 {
	return m.from
}

func (m *MappingEntry) To() uint32 {
	return m.from + m.length
}

func (m *MappingEntry) Length() uint32 {
	return m.length
}

func (m *MappingEntry) Name() string {
	return m.name
}

func (m *MappingEntry) SetName(name string) {
	m.name = name
}

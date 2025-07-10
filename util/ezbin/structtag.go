package ezbin

import (
	"reflect"
	"strconv"
)

// This is modified code from the GO standard library.
// It finds the position of a specific key/value pair in a tag.
func StructTagFind(tag reflect.StructTag, key string) (start int, end int, ok bool) {
	i := 0

	for tag != "" {
		// Skip leading space.
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		if tag[i:] == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		start := i
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == start || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[start:i])

		// Scan quoted string to find value.
		i += 2
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}

		i++
		if i > len(tag) {
			break
		}

		if key == name {
			return start, i, true
		}
	}

	// Tag not found :(
	return 0, 0, false
}

// Removes a key/value pair from the struct tag
func StructTagRemove(tag reflect.StructTag, key string) reflect.StructTag {
	start, end, ok := StructTagFind(tag, key)
	if !ok {
		return tag
	}

	return tag[:start] + tag[end:]
}

// Adds a key/value pair to the struct tag
func StructTagAdd(tag reflect.StructTag, key string, value string) reflect.StructTag {
	val := string(tag) + " " + key + ":" + strconv.Quote(value)
	return reflect.StructTag(val)
}

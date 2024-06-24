package pit

import (
	"bytes"
	"strings"
)

type PitString struct {
	text        []string
	escape      [][]byte
	escapeCount []int
}

// Returns the (non-formatted) text
// TODO: what text encoding for non-ascii characters is used?
func (s *PitString) String() string {
	b := strings.Builder{}
	escapePos := 0

	for i, escapes := range s.escapeCount {
		for ; escapes > 0; escapes-- {
			escape := s.escape[escapePos]
			escapePos++

			switch escape[0] {
			case '\n':
				b.WriteByte('\n')
			}
		}
		b.WriteString(s.text[i])
	}

	return b.String()
}

// Re-encodes a PitString to a binary, null-terminated blob
func (s *PitString) Encode() []byte {
	b := bytes.Buffer{}
	escapePos := 0

	for i, formats := range s.escapeCount {
		for ; formats > 0; formats-- {
			b.WriteByte(0xFF)
			b.Write(s.escape[escapePos])
			escapePos++
		}
		b.WriteString(s.text[i])
	}
	b.WriteByte(0)
	return b.Bytes()
}

// Decodes a Partners in Time string.
// Check the documentation for a rundown of the format.
// This function terminates at a null character, or when no more characters
// are available.
func DecodeString(str []byte) PitString {
	buf := bytes.NewReader(str)
	out := PitString{}
	strbuf := [256]byte{}

	for {
		// Collect escape sequences
		escapeCount := 0
		for {
			chr, _ := buf.ReadByte()
			if chr != 0xFF {
				buf.UnreadByte()
				out.escapeCount = append(out.escapeCount, escapeCount)
				break
			}

			escapeCount++
			chr, _ = buf.ReadByte()

			switch chr {
			default:
				out.escape = append(out.escape, []byte{chr})
			}
		}

		// Now read string for as long as we can
		strbuf := strbuf[:0]
		finish := false
		for {
			chr, _ := buf.ReadByte()
			if chr == 0xFF {
				buf.UnreadByte()
				break
			}
			if chr == 0x00 {
				finish = true
				break
			}
			strbuf = append(strbuf, chr)
		}
		out.text = append(out.text, string(strbuf))

		if finish {
			break
		}
	}

	return out
}

// Encodes a PitString back to a binary, null-terminated blob
func EncodeString(str PitString) []byte {
	return str.Encode()
}

// Builds a PitString from a regular string.
func NewString(str string) PitString {
	split := strings.Split(str, "\n")
	return DecodeString([]byte(strings.Join(split, "\xFF\n")))
}

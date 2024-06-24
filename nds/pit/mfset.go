package pit

import "encoding/binary"

// Get an array of strings from a MFset blob.
func DecodeMfset(mfset []byte, num int) []PitString {
	out := make([]PitString, num)
	for i := 0; i < num; i++ {
		offset := binary.LittleEndian.Uint32(mfset[i*4:])
		out[i] = DecodeString(mfset[offset:])
	}

	return out
}

// Get an array of grouped strings from a MFset blob.
// Useful for stuff like item names, where each item has 3 entries.
func DecodeMfsetGrouped(mfset []byte, groupSize int, groupNum int) [][]PitString {
	out := make([][]PitString, groupNum)
	strs := DecodeMfset(mfset, groupSize*groupNum)
	for i := range out {
		out[i] = strs[i*groupSize : i*groupSize+groupSize]
	}
	return out
}

func EncodeMfset(mfset []PitString) []byte {
	out := make([]byte, len(mfset)*4, 1024)
	for i := range mfset {
		binary.LittleEndian.PutUint32(out[i*4:], uint32(len(out)))
		out = append(out, mfset[i].Encode()...)
	}
	return out
}

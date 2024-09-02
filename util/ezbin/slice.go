package ezbin

import "bytes"

// Create a pre-initialized array.
func FillerArray[K any](length int, value K) []K {
	buf := make([]K, length)
	for i := range buf {
		buf[i] = value
	}
	return buf
}

func Put(buf []byte, pos int, data ...any) ([]byte, error) {

	// Write data to temp buffer
	tmp := &bytes.Buffer{}
	err := Write(tmp, data...)
	if err != nil {
		return nil, err
	}

	// Expand existing buffer if needed
	tempd := tmp.Bytes()
	if pos+len(tempd) > len(buf) {
		buf = append(buf, make([]byte, pos+len(tempd)-len(buf))...)
	}

	// Copy buffered data into real buffer
	copy(buf[pos:], tempd)
	return buf, nil
}

func PutAny[K any](buf []K, pos int, data ...K) []K {
	if len(buf) < pos+len(data) {
		buf = append(buf, make([]K, (pos+len(data))-len(buf))...)
	}

	// Copy buffered data into real buffer
	copy(buf[pos:], data)
	return buf
}

func Get(buf []byte, pos int, data ...any) error {
	return Read(bytes.NewReader(buf[pos:]), data...)
}

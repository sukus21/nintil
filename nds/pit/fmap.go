package pit

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func LoadMaps(r io.ReadSeeker) {
	oseek, _ := r.Seek(0, io.SeekCurrent)
	defer r.Seek(oseek, io.SeekStart)
	count := 0xBFA
	offsets := make([]uint32, 0, count)
	// lengths := make([]uint32, 0, count-1)

	// Read individual map offsets
	r.Seek(0, io.SeekStart)
	for i := 0; i < count; i++ {
		n := uint32(0)
		binary.Read(r, binary.LittleEndian, &n)
		offsets = append(offsets, n)
	}

	// Read map contents to new file
	os.MkdirAll("maps", 0666)
	for i := 0; i < count-1; i++ {
		r.Seek(int64(offsets[i]), io.SeekStart)
		buf := make([]byte, offsets[i+1]-offsets[i])
		r.Read(buf)
		os.WriteFile(fmt.Sprintf("maps/%d.map", i), buf, 0666)
	}
	fmt.Println("peepee")
}

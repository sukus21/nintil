package pit

import (
	"bytes"
	"encoding/binary"
	"image"
	"io"

	"github.com/sukus21/nintil/compression/rlz"
	"github.com/sukus21/nintil/nds"
)

func DecodeDebugMap(rom *nds.Rom, set int) image.PalettedImage {
	elem, _ := rom.Filesystem.Open("Etc/Sasaki/TestMapData.dat")
	raw, _ := io.ReadAll(elem)
	dat := UnpackDat(raw)
	tlm, _ := rlz.Decompress(dat[set*3+2])

	// Create tile graphics
	tileData, _ := rlz.Decompress(dat[set*3+1])
	tiles := nds.DeserializeTiles8BPP(tileData)
	palette := nds.DeserializePalette(dat[set*3+0], true)

	// Build tilemap and return
	mapWidth := 64
	img := nds.NewTilemap(mapWidth, len(tlm)/mapWidth*2, tiles, palette)
	binary.Read(bytes.NewReader(tlm), binary.LittleEndian, img.Attributes)
	return img
}

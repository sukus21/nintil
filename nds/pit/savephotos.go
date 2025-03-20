package pit

import (
	"bytes"
	"encoding/binary"
	"image"
	"io"

	"github.com/sukus21/nintil/compression/lz10"
	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
)

func DecodeSavePhoto(rom *nds.Rom, id int) image.PalettedImage {
	elem, _ := rom.Filesystem.Open("SavePoint/SavePhoto.dat")
	b, _ := io.ReadAll(elem)
	elem.Close()

	tiles := nds.DeserializeTiles8BPP(util.Must1(lz10.Decompress(UnpackDatSingle(b, id*3+0))))
	tilemap := util.Must1(lz10.NewReader(bytes.NewReader(UnpackDatSingle(b, id*3+1))))
	palette := nds.DeserializePalette(UnpackDatSingle(b, id*3+2), true)

	// Create new image
	img := nds.NewTilemap(32, 24, tiles, palette)
	binary.Read(tilemap, binary.LittleEndian, img.Attributes)
	return img
}

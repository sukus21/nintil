package pit

import (
	"bytes"
	"image"
	"io"

	"github.com/sukus21/nintil/compression/lz10"
	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

func DecodeSavePhoto(rom *nds.Rom, id int) image.PalettedImage {
	elem, _ := rom.Filesystem.Open("SavePoint/SavePhoto.dat")
	b, _ := io.ReadAll(elem)
	elem.Close()

	tiles := nds.DeserializeTiles8BPP(util.Must1(lz10.Decompress(UnpackDatSingle(b, id*3+0))))
	tilemap := util.Must1(lz10.NewReader(bytes.NewReader(UnpackDatSingle(b, id*3+1))))
	palette := nds.DeserializePalette(UnpackDatSingle(b, id*3+2), true)

	// Create new image
	img := image.NewPaletted(
		image.Rect(0, 0, 256, 192),
		palette,
	)

	// Draw tiles onto image
	for i := 0; i < 32*24; i++ {
		raw := ezbin.ReadSingle[uint16](tilemap)
		tileId := raw & 0x3FF
		tileMirror := raw&0x0400 != 0
		tileFlip := raw&0x0800 != 0
		// tilePal := int(raw >> 12)

		tile := tiles[tileId]
		tile.Palette = palette
		x := (i & 0x1F) * 8
		y := (i >> 5) * 8
		nds.DrawTileSamePalette(
			img, tile,
			x, y,
			tileMirror, tileFlip,
		)
	}

	return img
}

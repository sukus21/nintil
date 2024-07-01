package pit

import (
	"encoding/binary"
	"image"
	"io"
	"math/bits"

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

	// Construct image
	imageWidth := 512
	mapWidth := imageWidth / 8
	imageHeight := len(tlm) / mapWidth * 4
	img := image.NewPaletted(
		image.Rect(0, 0, imageWidth, imageHeight),
		palette,
	)

	// Read tiles
	for i := 0; i < len(tlm); i += 2 {
		raw := binary.LittleEndian.Uint16(tlm[i : i+2])
		tileId := raw & 0x03FF
		tileMirror := raw&0x0400 != 0
		tileFlip := raw&0x0800 != 0
		// paletteOffset := (raw >> 8) & 0xF0

		tile := tiles[tileId]
		tile.Palette = palette // [paletteOffset : paletteOffset+16]

		mask := mapWidth - 1
		x := ((i / 2) & (mask)) * 8
		y := ((i / 2) >> bits.OnesCount(uint(mask))) * 8
		nds.DrawTileSamePalette(
			img,
			tile,
			x,
			y,
			tileMirror,
			tileFlip,
		)
	}

	return img
}

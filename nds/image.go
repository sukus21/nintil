package nds

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// Implements image.PalettedImage
type Tile struct {
	pix [64]byte
	color.Palette
}

func (t *Tile) ColorIndexAt(x int, y int) byte {
	return t.pix[x+y<<3]
}

// Why do you need this?
func (t *Tile) ColorModel() color.Model {
	return t.Palette
}

// Always starts at 0,0 and ends at 8,8
func (t *Tile) Bounds() image.Rectangle {
	return image.Rect(0, 0, 8, 8)
}

// Palette lookup
func (t *Tile) At(x int, y int) color.Color {
	return t.Palette[t.ColorIndexAt(x, y)]
}

// Read RGB555 palette from raw data.
func DeserializePalette(b []byte, firstTransparent bool) color.Palette {
	pal := make(color.Palette, len(b)/2)
	for i := 0; i < len(pal); i++ {
		c := (uint16(b[i*2+1]) << 8) | uint16(b[i*2+0])
		a := byte(255)
		if i == 0 && firstTransparent {
			a = 0
		}
		cr := (byte(c) & 0x1F)
		cg := (byte(c>>5) & 0x1F)
		cb := (byte(c>>10) & 0x1F)
		p := color.RGBA{
			R: cr<<3 | cr>>2,
			G: cg<<3 | cg>>2,
			B: cb<<3 | cb>>2,
			A: a,
		}
		pal[i] = p
	}
	return pal
}

// Serialized specified palette to raw RGB555 palette data.
// Pad specifies how long the file should be.
// Set pad to 0 to ignore.
// Will throw an error if (nonzero) pad is smaller than output size.
func SerializePalette(src color.Palette, pad int) ([]byte, error) {

	// Apply padding
	length := len(src) * 2
	if pad > 0 {
		if length > pad {
			return nil, fmt.Errorf("serialize palette: padding should be >= final output")
		}
		length = pad
	}
	cols := make([]byte, length)

	// Serialize palette
	for i, v := range src {
		if v == nil {
			continue
		}
		col := color.RGBAModel.Convert(v).(color.RGBA)
		r := uint16(col.R) >> 3
		g := uint16(col.G) >> 3
		b := uint16(col.B) >> 3
		binary.LittleEndian.PutUint16(cols[i*2:], r|(g<<5)|(b<<10))
	}

	// Return
	return cols, nil
}

// Reads a list of 8x8, 4BPP tiles from raw data.
// Tiles will not have a palette associated when returned.
func DeserializeTiles4BPP(b []byte) []*Tile {
	tiles := make([]*Tile, len(b)/32)
	for i := range tiles {
		dat := [64]byte{}
		for j := 0; j < 32; j++ {
			r := b[i*32+j]
			dat[j*2+0] = r & 0x0F
			dat[j*2+1] = r >> 4
		}
		tiles[i] = &Tile{pix: dat}
	}
	return tiles
}

// Throws an error if color indexes go above 15.
func SerializeTiles4BPP(tiles []*Tile) ([]byte, error) {
	buf := make([]byte, len(tiles)*32)
	for i, v := range tiles {
		for j := 0; j < 64; j += 2 {
			if v.pix[j+0] > 15 || v.pix[j+1] > 15 {
				return nil, fmt.Errorf("serialize tiles 4bpp: palette index can't be above 15")
			}
			buf[i*32+j/2] = v.pix[j+0] | v.pix[j+1]<<4
		}
	}
	return buf, nil
}

// Reads a list of 8x8, 8BPP tiles from raw data.
// Tiles will not have a palette associated when returned.
func DeserializeTiles8BPP(b []byte) []*Tile {
	tiles := make([]*Tile, len(b)/64)
	for i := range tiles {
		tiles[i] = &Tile{pix: [64]byte{}}
		copy(tiles[i].pix[:], b[i*64:i*64+64])
	}
	return tiles
}

// Error is always nil and can be ignored.
func SerializeTiles8BPP(tiles []*Tile) ([]byte, error) {
	buf := make([]byte, len(tiles)*64)
	for i, v := range tiles {
		copy(buf[i*64:], v.pix[:])
	}
	return buf, nil
}

// Draw a single tile onto a canvas
func DrawTile(canvas draw.Image, tile *Tile, x int, y int, mirror bool, flip bool) {
	for i := 0; i < 8; i++ {
		sx := i
		if mirror {
			sx = 7 - i
		}
		for j := 0; j < 8; j++ {
			sy := j
			if flip {
				sy = 7 - j
			}
			pix := tile.At(sx, sy)
			canvas.Set(x+i, y+j, pix)
		}
	}
}

// Same as DrawTile, but can skip a lot of code to optimize
func DrawTileSamePalette(canvas *image.Paletted, tile *Tile, x int, y int, mirror bool, flip bool) {
	for i := 0; i < 8; i++ {
		sx := i
		if mirror {
			sx = 7 - i
		}
		for j := 0; j < 8; j++ {
			sy := j
			if flip {
				sy = 7 - j
			}
			pix := tile.ColorIndexAt(sx, sy)
			canvas.SetColorIndex(x+i, y+j, pix)
		}
	}
}

// Same as DrawTileSamePalette, but can shift palettes
func DrawTileShiftPalette(canvas *image.Paletted, tile *Tile, x int, y int, mirror bool, flip bool, palshift int) {
	for i := range 8 {
		sx := i
		if mirror {
			sx = 7 - i
		}
		for j := range 8 {
			sy := j
			if flip {
				sy = 7 - j
			}
			pix := tile.ColorIndexAt(sx, sy)
			if pix == 0 {
				canvas.SetColorIndex(x+i, y+j, 0)
			} else {
				canvas.SetColorIndex(x+i, y+j, pix+byte(palshift))
			}
		}
	}
}

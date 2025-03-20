package nds

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"

	"github.com/sukus21/nintil/util/ezbin"
)

// A single 8x8 tile
type Tile struct {
	Pix [64]byte
}

func (t *Tile) ColorIndexAt(x int, y int) byte {
	return t.Pix[x+y<<3]
}

// Read RGB555 palette from raw data.
func ReadPalette(r io.Reader, numColors int, firstTransparent bool) (color.Palette, error) {
	buf := make([]byte, numColors*2)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	return DeserializePalette(buf, firstTransparent), nil
}

// Read RGB555 palette from raw data.
func DeserializePalette(b []byte, firstTransparent bool) color.Palette {
	pal := make(color.Palette, len(b)/2)
	for i := range len(pal) {
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
func DeserializeTiles4BPP(b []byte) []Tile {
	tiles := make([]Tile, len(b)/32)
	for i := range tiles {
		tile := &tiles[i]
		for j := 0; j < 32; j++ {
			r := b[i*32+j]
			tile.Pix[j*2+0] = r & 0x0F
			tile.Pix[j*2+1] = r >> 4
		}
	}
	return tiles
}

// Throws an error if color indexes go above 15.
func SerializeTiles4BPP(tiles []Tile) ([]byte, error) {
	buf := make([]byte, len(tiles)*32)
	for i, v := range tiles {
		for j := 0; j < 64; j += 2 {
			if v.Pix[j+0] > 15 || v.Pix[j+1] > 15 {
				return nil, fmt.Errorf("serialize tiles 4bpp: palette index can't be above 15")
			}
			buf[i*32+j/2] = v.Pix[j+0] | v.Pix[j+1]<<4
		}
	}
	return buf, nil
}

// Reads a list of 8x8, 8BPP tiles from raw data.
func DeserializeTiles8BPP(b []byte) []Tile {
	tiles := make([]Tile, len(b)/64)
	for i := range tiles {
		copy(tiles[i].Pix[:], b[i*64:i*64+64])
	}
	return tiles
}

// Error is always nil and can be ignored.
func SerializeTiles8BPP(tiles []*Tile) ([]byte, error) {
	buf := make([]byte, len(tiles)*64)
	for i, v := range tiles {
		copy(buf[i*64:], v.Pix[:])
	}
	return buf, nil
}

// Draw a single tile onto a canvas
func DrawTile(canvas draw.Image, palette color.Palette, tile *Tile, x int, y int, mirror bool, flip bool) {
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
			canvas.Set(x+i, y+j, palette[pix])
		}
	}
}

// Same as DrawTile, but can skip a lot of code to optimize
func DrawTileSamePalette(canvas *image.Paletted, tile *Tile, x int, y int, mirror bool, flip bool) {
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

// A bitfield, but can be used like a regular uint16.
type TilemapAttributes uint16

// Tile index
func (attr TilemapAttributes) GetTileIndex() int {
	return ezbin.Bitget[int](attr, 10, 0)
}
func (attr *TilemapAttributes) SetTileIndex(tileIndex int) {
	*attr = ezbin.Bitset(*attr, tileIndex, 10, 0)
}

// If this tile should be mirrored or not
func (attr TilemapAttributes) GetFlipX() bool {
	return ezbin.BitgetFlag(attr, 10)
}
func (attr *TilemapAttributes) SetFlipX(flipX bool) {
	*attr = ezbin.BitsetFlag(*attr, flipX, 10)
}

// If this tile should be flipped or not
func (attr TilemapAttributes) GetFlipY() bool {
	return ezbin.BitgetFlag(attr, 11)
}
func (attr *TilemapAttributes) SetFlipY(flipY bool) {
	*attr = ezbin.BitsetFlag(*attr, flipY, 11)
}

// For 4bpp tiles, decide which set of 16 colors to use from a 256 color palette.
// Value ranges from from 0 to 15.
func (attr TilemapAttributes) GetPaletteShift() int {
	return ezbin.Bitget[int](attr, 4, 12)
}
func (attr *TilemapAttributes) SetPaletteShift(paletteShift int) {
	*attr = ezbin.Bitset(*attr, paletteShift, 4, 12)
}

// A tilemap.
// Implements image.PalettedImage
type Tilemap struct {
	Palette    color.Palette
	Tileset    []Tile
	Attributes []TilemapAttributes

	// Size in tiles
	width, height int
}

func NewTilemap(width, height int, tileset []Tile, palette color.Palette) *Tilemap {
	return &Tilemap{
		Palette:    palette,
		Tileset:    tileset,
		width:      width,
		height:     height,
		Attributes: make([]TilemapAttributes, width*height),
	}
}

func (tilemap *Tilemap) ColorModel() color.Model {
	return tilemap.Palette[:min(256, len(tilemap.Palette))]
}

func (tilemap *Tilemap) Bounds() image.Rectangle {
	return image.Rect(0, 0, tilemap.width*8, tilemap.height*8)
}

func (tilemap *Tilemap) At(x, y int) color.Color {
	pixelX, pixelY, attributes := tilemap.getEntryAt(x, y)

	// Get tile color
	tile := &tilemap.Tileset[attributes.GetTileIndex()]
	colorIndex := tile.ColorIndexAt(pixelX, pixelY)
	if colorIndex == 0 {
		_, _, _, a := tilemap.Palette[0].RGBA()
		if a == 0 {
			return color.Transparent
		}
	}

	// Modify palette index
	colorIndex += uint8(attributes.GetPaletteShift() * 16)
	return tilemap.Palette[colorIndex]
}

func (tilemap *Tilemap) ColorIndexAt(x, y int) uint8 {
	pixelX, pixelY, attributes := tilemap.getEntryAt(x, y)

	// Get tile color
	tile := &tilemap.Tileset[attributes.GetTileIndex()]
	colorIndex := tile.ColorIndexAt(pixelX, pixelY)
	if colorIndex == 0 {
		return 0
	}

	// Modify palette index
	return colorIndex + uint8(attributes.GetPaletteShift()*16)
}

func (tilemap *Tilemap) getEntryAt(x, y int) (int, int, TilemapAttributes) {
	// Get tilemap entry
	tileX := x / 8
	tileY := y / 8
	attributes := tilemap.Attributes[tileY*tilemap.width+tileX]

	// Get tile pixel
	pixelX := x % 8
	if attributes.GetFlipX() {
		pixelX = 7 - pixelX
	}
	pixelY := y % 8
	if attributes.GetFlipY() {
		pixelY = 7 - pixelY
	}

	return pixelX, pixelY, attributes
}

func (tilemap *Tilemap) GetAttributes(x, y int) TilemapAttributes {
	return tilemap.Attributes[x+y*tilemap.width]
}

func (tilemap *Tilemap) SetAttributes(x, y int, attributes TilemapAttributes) {
	tilemap.Attributes[x+y*tilemap.width] = attributes
}

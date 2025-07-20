package nds

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"

	"github.com/sukus21/nintil/util/ezbin"
)

type TeximageParams uint32

func (b TeximageParams) GetOffset() int            { return ezbin.Bitget[int](b, 16, 0) }
func (b TeximageParams) GetWidth() int             { return 8 << ezbin.Bitget[int](b, 3, 20) }
func (b TeximageParams) GetHeight() int            { return 8 << ezbin.Bitget[int](b, 3, 23) }
func (b TeximageParams) GetFormat() TextureFormat  { return ezbin.Bitget[TextureFormat](b, 3, 26) }
func (b TeximageParams) IsColor0Transparent() bool { return ezbin.BitgetFlag(b, 29) }

type TextureFormat int

const (
	TextureFormat_Invalid = TextureFormat(iota)
	TextureFormat_A3I5
	TextureFormat_2BPP
	TextureFormat_4BPP
	TextureFormat_8BPP
	TextureFormat_4x4Texel
	TextureFormat_A5I3
	TextureFormat_RGBA5551
)

func (f TextureFormat) PaletteSize() int {
	switch f {
	case TextureFormat_2BPP:
		return 4
	case TextureFormat_4BPP:
		return 16
	case TextureFormat_8BPP:
		return 256
	case TextureFormat_A3I5:
		return 32
	case TextureFormat_A5I3:
		return 8
	case TextureFormat_4x4Texel:
		return 4 // Is this correct? According to GBAtek, it might be correct???
	case TextureFormat_RGBA5551:
		return 0

	// Invalid texture format
	default:
		return 0
	}
}

// Palette can be nil, if texture is RGBA5551
func ReadTexture(r io.Reader, params TeximageParams, palette color.Palette) (image.Image, error) {

	// If first color is transparent, duplicate palette and force it to be
	if params.IsColor0Transparent() && len(palette) > 0 {
		_, _, _, a := palette[0].RGBA()
		if a != 0 {
			palette = append(make(color.Palette, 0, len(palette)), palette...)
			palette[0] = color.Alpha{A: 0}
		}
	}

	switch params.GetFormat() {
	case TextureFormat_2BPP:
		return ReadTexture2BPP(r, params.GetWidth(), params.GetHeight(), palette)
	case TextureFormat_4BPP:
		return ReadTexture4BPP(r, params.GetWidth(), params.GetHeight(), palette)
	case TextureFormat_8BPP:
		return ReadTexture8BPP(r, params.GetWidth(), params.GetHeight(), palette)
	case TextureFormat_A3I5:
		return ReadTextureA3I5(r, params.GetWidth(), params.GetHeight(), palette)
	case TextureFormat_A5I3:
		return ReadTextureA5I3(r, params.GetWidth(), params.GetHeight(), palette)
	case TextureFormat_4x4Texel:
		return nil, fmt.Errorf("cannot stream 4x4 texel compressed texture, requires 2 streams")
	case TextureFormat_RGBA5551:
		return ReadTextureRGBA5551(r, params.GetWidth(), params.GetHeight())

	default:
		return nil, fmt.Errorf("invalid texture format %d", params.GetFormat())
	}
}

func ReadTexture2BPP(r io.Reader, width int, height int, palette color.Palette) (*image.Paletted, error) {

	// Read source data
	buf := make([]byte, (width*height)/4)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	// Deserialize data
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	i := 0
	for _, pix := range buf {
		img.Pix[i] = (pix >> 0) & 0x03
		i++
		img.Pix[i] = (pix >> 2) & 0x03
		i++
		img.Pix[i] = (pix >> 4) & 0x03
		i++
		img.Pix[i] = (pix >> 6) & 0x03
		i++
	}

	return img, nil
}

func ReadTexture4BPP(r io.Reader, width int, height int, palette color.Palette) (*image.Paletted, error) {

	// Read source data
	buf := make([]byte, (width*height)/2)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	// Deserialize data
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	i := 0
	for _, pix := range buf {
		img.Pix[i] = (pix >> 0) & 0x0F
		i++
		img.Pix[i] = (pix >> 4) & 0x0F
		i++
	}

	return img, nil
}

func ReadTexture8BPP(r io.Reader, width int, height int, palette color.Palette) (*image.Paletted, error) {
	// Deserialize data
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	_, err := io.ReadFull(r, img.Pix)
	return img, err
}

func buildAlphaPalette(oldPalette color.Palette, alpha int, index int) color.Palette {
	numAlpha := 1 << alpha
	numIndex := 1 << index

	// Modify palette
	scale := float64(1) / float64(numAlpha-1)
	newPalette := make(color.Palette, 0, 0x100)
	for i := range numIndex {
		var oldColor color.Color = color.Black
		if i < len(oldPalette) {
			oldColor = oldPalette[i]
		}

		r, g, b, _ := oldColor.RGBA()

		newColor := color.RGBA64{R: uint16(r), G: uint16(g), B: uint16(b)}
		for a := range numAlpha {
			newColor.A = uint16(scale * float64(a) * 0x10000)
			newPalette[i+a*numIndex] = newColor
		}
	}

	return newPalette
}

// This is just 8BPP but with a silly palette
func ReadTextureA3I5(r io.Reader, width int, height int, palette color.Palette) (*image.Paletted, error) {
	newPalette := buildAlphaPalette(palette, 3, 5)
	return ReadTexture8BPP(r, width, height, newPalette)
}

// This is just 8BPP but with a silly palette
func ReadTextureA5I3(r io.Reader, width int, height int, palette color.Palette) (*image.Paletted, error) {
	newPalette := buildAlphaPalette(palette, 5, 3)
	return ReadTexture8BPP(r, width, height, newPalette)
}

func ReadTextureRGBA5551(r io.Reader, width int, height int) (*image.RGBA, error) {
	buf := make([]uint16, width*height)
	err := binary.Read(r, binary.LittleEndian, buf)
	if err != nil {
		return nil, err
	}

	// Write color info to texture
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	i := 0
	for _, pix := range buf {
		img.Pix[i+0] = uint8(pix>>0) & 0x1F
		img.Pix[i+1] = uint8(pix>>5) & 0x1F
		img.Pix[i+2] = uint8(pix>>10) & 0x1F
		img.Pix[i+3] = 0
		if pix&0x8000 != 0 {
			img.Pix[i+3] = 0xFF
		}

		i += 4
	}

	return img, nil
}

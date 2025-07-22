package g3d

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

type SubfileTEX0 struct {
	texBlockRaw         []byte
	texBlockCompressed1 []uint32
	texBlockCompressed2 []uint16
	palBlock            []byte

	Textures []TexTexture
	Palettes []TexPalette
}

func (f *SubfileTEX0) GetStamp() string {
	return "TEX0"
}

func (f *SubfileTEX0) EzbinDecode(r io.Reader) (err error) {
	defer util.Recover(&err)
	rs, ok := r.(io.ReadSeeker)
	if !ok {
		return ezbin.ErrNotSeeker
	}

	// Read binary structure
	stub := util.Must1(ezbin.Decode[struct {
		Offset         int64  `ezbin_tell:"tex"`
		Stamp          string `ezbin_string:"ascii,4"`
		SectionSize    uint32
		_              uint32
		TexBlockLength uint16
		Textures       NameListImmediate[struct {
			TeximageParams nds.TeximageParams
			_              uint32
		}] `ezbin_offset:"u16,tex"`
		_                      uint32
		TexBlockOffset         uint32
		_                      uint32
		CompressedBlockLen     uint16
		CompressedInfoOffset   uint16
		_                      uint32
		Compressed1BlockOffset uint32
		Compressed2BlockOffset uint32
		_                      uint32
		PalBlockLength         uint16
		_                      uint16
		Palettes               NameListImmediate[struct {
			Offset uint16
			_      uint16
		}] `ezbin_offset:"u32,tex"`
		PalBlockOffset uint32
	}](rs))

	// Try to figure out what any of this means
	f.texBlockRaw = make([]byte, stub.TexBlockLength<<3)
	util.Must1(rs.Seek(int64(stub.TexBlockOffset)+stub.Offset, io.SeekStart))
	util.Must1(io.ReadFull(rs, f.texBlockRaw))

	// Try to figure out what any of this means
	f.texBlockCompressed1 = make([]uint32, stub.CompressedBlockLen<<1)
	util.Must1(rs.Seek(int64(stub.Compressed1BlockOffset)+stub.Offset, io.SeekStart))
	util.Must(binary.Read(rs, binary.LittleEndian, f.texBlockCompressed1))

	// Try to figure out what any of this means
	f.texBlockCompressed2 = make([]uint16, stub.CompressedBlockLen<<1)
	util.Must1(rs.Seek(int64(stub.Compressed2BlockOffset)+stub.Offset, io.SeekStart))
	util.Must(binary.Read(rs, binary.LittleEndian, f.texBlockCompressed2))

	// Try to figure out what any of this means
	f.palBlock = make([]byte, stub.PalBlockLength<<3)
	util.Must1(rs.Seek(int64(stub.PalBlockOffset)+stub.Offset, io.SeekStart))
	util.Must1(io.ReadFull(rs, f.palBlock))

	// Copy textures over
	f.Textures = make([]TexTexture, len(stub.Textures.Data))
	for i := range f.Textures {
		f.Textures[i].TeximageParams = stub.Textures.Data[i].TeximageParams
		f.Textures[i].Name = stub.Textures.Names[i]
	}

	// Copy palettes over
	f.Palettes = make([]TexPalette, len(stub.Palettes.Data))
	for i := range f.Palettes {
		f.Palettes[i].Offset = stub.Palettes.Data[i].Offset
		f.Palettes[i].Name = stub.Palettes.Names[i]
	}

	return nil
}

func (f *SubfileTEX0) GetTextureFormat(index int) nds.TextureFormat {
	return f.Textures[index].TeximageParams.GetFormat()
}

func (f *SubfileTEX0) GetTexture(index int, palette color.Palette) (image.Image, error) {
	tex := f.Textures[index]
	offset := int(tex.TeximageParams.GetOffset())

	if tex.TeximageParams.GetFormat() == nds.TextureFormat_4x4Texel {
		return nil, fmt.Errorf("TODO: 4x4 texel decompression")
	}

	r := bytes.NewReader(f.texBlockRaw[offset<<3:])
	return nds.ReadTexture(r, tex.TeximageParams, palette)
}

func (f *SubfileTEX0) GetPalette(index int, format nds.TextureFormat) color.Palette {
	paletteInfo := f.Palettes[index]

	palLength := format.PaletteSize()

	offset := int(paletteInfo.Offset) << 3
	buf := f.palBlock[offset:min(offset+palLength*2, len(f.palBlock))]

	return nds.DeserializePalette(buf, false)
}

type TexPalette struct {
	Offset uint16
	Name   string
}

type TexTexture struct {
	TeximageParams nds.TeximageParams
	Name           string
}

package pit

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"io"

	"github.com/sukus21/nintil/compression/rlz"
	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util/ezbin"
)

// Information about which files belong together.
// Embedded somewhere in the ARM9 binary, for some reason...
type FMapInfo struct {
	Tilesets [3]uint32
	Bundle   uint32
	Unknown  uint32
}

// Information about the given map.
// Mustly unknown...
type FMapBundle struct {
	Tilemaps [3][]byte
	Palettes [3][]byte
	Metadata FMapMetadata
	Unknown  [6][]byte
}

type FMapMetadata struct {
	// Width of map in tiles
	Width uint16

	// Height of map in tiles
	Height uint16

	// Unknown.
	// Often contains 0xFF?
	Unknown1 uint8

	// Bottom 3 bits incicate if a layer is 8BPP (1) or 4BPP (0).
	// The rest is unknown.
	LayerBitdepths uint8

	Unknown2 [6]byte
}

// Implementation of FMapReader
type FMapReader struct {
	fmapInfo []FMapInfo
	fmapFile io.ReaderAt
}

// The given FMapInfo blob is invalid
var ErrInvalidFMapInfo = errors.New("FMap: invalid FMap info block size")

// The bundle file for a map is invalid
var ErrInvalidFMapBundle = errors.New("FMap: invalid FMap bundle file")

// fmapInfoBlock is the block stored somewhere in the ARM9 binary.
// It's super annoying that it's not stored in the NitroFS, but what can we do...
//
// fmapFile is the "FMap/FMapData.dat" file
func NewFMapReader(fmapInfoBlock []byte, fmapFile io.ReaderAt) (*FMapReader, error) {
	if len(fmapInfoBlock)%(5*4) != 0 {
		return nil, ErrInvalidFMapInfo
	}

	// Deserialize FMmapInfo elements
	fmapMaps := len(fmapInfoBlock) / 20
	fmapInfos := make([]FMapInfo, fmapMaps)
	err := ezbin.Read(bytes.NewReader(fmapInfoBlock), fmapInfos)
	if err != nil {
		return nil, err
	}

	// Return reader
	return &FMapReader{
		fmapInfo: fmapInfos,
		fmapFile: fmapFile,
	}, nil
}

// Number of total maps
func (r *FMapReader) MapCount() int {
	return len(r.fmapInfo)
}

func (r *FMapReader) readFMapBundleDats(bundleFile uint32) ([][]byte, error) {
	bundleData, err := r.decompressFile(bundleFile)
	if err != nil {
		return nil, err
	}
	if !IsDat(bundleData) {
		return nil, ErrNotDat
	}

	return UnpackDat(bundleData), nil
}

func (r *FMapReader) decompressFile(fileId uint32) ([]byte, error) {
	fileReader, err := OpenDat(r.fmapFile, int(fileId))
	if err != nil {
		return nil, err
	}

	decompressor, err := rlz.NewReader(fileReader)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(decompressor)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (r *FMapReader) NewRenderer(mapId int) (*FMapRenderer, error) {
	info := r.fmapInfo[mapId]

	// Open bundle file
	bundleData, err := r.readFMapBundleDats(info.Bundle)
	if err == ErrNotDat || len(bundleData) != 13 {
		return nil, ErrInvalidFMapBundle
	}

	// Read bundle
	bundle := FMapBundle{}
	copy(bundle.Tilemaps[:], bundleData[0:])
	copy(bundle.Palettes[:], bundleData[3:])
	bundle.Metadata = ezbin.ReadSingle[FMapMetadata](bytes.NewReader(bundleData[6]))
	copy(bundle.Unknown[:], bundleData[7:])

	// Return renderer
	return &FMapRenderer{
		r:      r,
		Info:   info,
		Bundle: bundle,
	}, nil
}

type FMapRenderer struct {
	r          *FMapReader
	Info       FMapInfo
	Bundle     FMapBundle
	layerCache [3]*image.Paletted
}

// Renders the full tilemap, with all 3 layers.
func (r *FMapRenderer) RenderMap() (image.Image, error) {
	// Read room width and height
	width := int(r.Bundle.Metadata.Width)
	height := int(r.Bundle.Metadata.Height)
	imgRect := image.Rect(0, 0, width*8, height*8)

	// Create final output image
	final := image.NewRGBA(imgRect)
	zp := image.Pt(0, 0)
	for i := range 3 {
		layerId := 2 - i
		if r.HasLayer(layerId) {
			layer, err := r.RenderLayer(layerId)
			if err != nil {
				return nil, err
			}
			draw.DrawMask(final, imgRect, layer, zp, layer, zp, draw.Over)
		}
	}

	// Done!
	return final, nil
}

// Renders a single layer.
// Output image may not be valid, is the layer does not exist (use HasLayer).
func (r *FMapRenderer) RenderLayer(layerId int) (image.Image, error) {
	if r.layerCache[layerId] != nil {
		return r.layerCache[layerId], nil
	}

	// Read room width and height
	width := int(r.Bundle.Metadata.Width)
	height := int(r.Bundle.Metadata.Height)
	imgRect := image.Rect(0, 0, width*8, height*8)

	// Read palette and create image
	palette := nds.DeserializePalette(r.Bundle.Palettes[layerId], layerId != 2)
	if len(palette) > 256 {
		palette = palette[:256]
	}
	img := image.NewPaletted(imgRect, palette)

	// Read tileset
	var tileset []*nds.Tile
	tilesetBytes, err := r.r.decompressFile(r.Info.Tilesets[layerId])
	if err != nil {
		return nil, err
	}

	// 4BPP or 8BPP
	props := r.Bundle.Metadata.LayerBitdepths
	if props&(1<<layerId) == 0 {
		tileset = nds.DeserializeTiles4BPP(tilesetBytes)
	} else {
		tileset = nds.DeserializeTiles8BPP(tilesetBytes)
	}

	// Render tiles to image
	tlm := bytes.NewReader(r.Bundle.Tilemaps[layerId])
	for j := 0; tlm.Len() != 0; j++ {
		posX := j % width
		posY := j / width

		// Get properties and draw tile
		tileData := ezbin.ReadSingle[uint16](tlm)
		flipX := tileData&(1<<10) != 0
		flipY := tileData&(1<<11) != 0
		paletteShift := tileData >> 12
		tileId := tileData & (0x3FF)
		nds.DrawTileShiftPalette(
			img, tileset[tileId],
			posX*8, posY*8,
			flipX, flipY,
			int(paletteShift)*16,
		)
	}

	// All done
	r.layerCache[layerId] = img
	return img, nil
}

func (r *FMapRenderer) HasLayer(layerId int) bool {
	if layerId < 0 || layerId > 2 {
		return false
	}
	return r.Info.Tilesets[layerId] != 0xFFFF_FFFF
}

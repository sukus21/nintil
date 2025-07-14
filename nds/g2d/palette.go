package g2d

import (
	"bytes"
	"fmt"
	"image/color"
	"io"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

type NCLR struct {
	Colors         color.Palette
	PaletteIndices []uint16
}

// Builds palette from colors and index data.
// If index data is empty, simply returns the colors array.
func (nclr *NCLR) GetPalette() color.Palette {
	if len(nclr.PaletteIndices) == 0 {
		return nclr.Colors
	}

	out := make(color.Palette, len(nclr.PaletteIndices))
	for i := range out {
		out[i] = nclr.Colors[nclr.PaletteIndices[i]]
	}
	return out
}

type blockPLTT struct {
	ColorFormat     uint32
	ExtentedPalette uint32
	PaletteSize     uint32
	PaletteData     []byte `ezbin_offset:"u32" ezbin_length:"PaletteSize"`
}

type blockPCMP struct {
	NumPalettes    uint16
	_              uint16
	PaletteIndices []uint16 `ezbin_offset:"u16" ezbin_length:"NumPalettes"`
}

func ReadNCLR(r io.ReadSeeker) (_ *NCLR, err error) {
	defer util.Recover(&err)
	out := new(NCLR)

	g2d := util.Must1(ezbin.Decode[G2DFile](r))
	for i := range g2d.Blocks {
		block := &g2d.Blocks[i]
		br := bytes.NewReader(block.Data)

		switch block.Stamp {
		case "TTLP": // PLTT
			pltt := util.Must1(ezbin.Decode[blockPLTT](br))
			out.Colors = nds.DeserializePalette(pltt.PaletteData, false)

		case "PMCP": // PCMP
			pcmp := util.Must1(ezbin.Decode[blockPCMP](br))
			out.PaletteIndices = pcmp.PaletteIndices

		default:
			return nil, fmt.Errorf("NCLR: invalid block type: %q", block.Stamp)
		}
	}

	return out, nil
}

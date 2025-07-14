package g2d

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

type NCBR struct {
	Char image.PalettedImage
	Cpos blockCPOS
}

type blockCHAR struct {
	Height       uint16
	Width        uint16
	ColorFormat  uint32
	MappingMode  uint32
	GraphicsType uint32
	GraphicsSize uint32
	GraphicsData []byte `ezbin_offset:"u32" ezbin_length:"GraphicsSize"`
}

type blockCPOS struct {
	X      uint16
	Y      uint16
	Width  uint16
	Height uint16
}

func ReadNCBR(r io.ReadSeeker, palette color.Palette) (_ *NCBR, err error) {
	defer util.Recover(&err)
	out := new(NCBR)

	g2d := util.Must1(ezbin.Decode[G2DFile](r))
	for i := range g2d.Blocks {
		block := &g2d.Blocks[i]
		br := bytes.NewReader(block.Data)

		switch block.Stamp {
		case "RAHC": // CHAR
			char := util.Must1(ezbin.Decode[blockCHAR](br))

			// Create output image
			img := image.NewPaletted(image.Rect(0, 0, int(char.Width)*8, int(char.Height)*8), palette)
			i := 0

			// Decode image, either 4bpp or 8bpp
			switch char.ColorFormat {
			case 3:
				for _, pix := range char.GraphicsData {
					img.Pix[i] = pix & 15
					i++
					img.Pix[i] = pix >> 4
					i++
				}
			case 4:
				for _, pix := range char.GraphicsData {
					img.Pix[i] = pix
					i++
				}
			default:
				return nil, fmt.Errorf("NCBR: invalid color format")
			}

			out.Char = img

		case "SOPC": // COPS
			out.Cpos = util.Must1(ezbin.Decode[blockCPOS](br))

		default:
			return nil, fmt.Errorf("NCBR: invalid block type: %q", block.Stamp)
		}
	}

	return out, nil
}

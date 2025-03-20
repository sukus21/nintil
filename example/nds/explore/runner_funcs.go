package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/nds/pit"
	"github.com/sukus21/nintil/util"
)

func addNdsFuncs(r *Runner) {
	// Open NDS ROM
	r.AddFunc("nds", func(r *Runner) any {
		romFile := ReadVarIf[*os.File](r, "rom_file")
		if romFile != nil {
			romFile.Close()
		}

		path := ReadArg[string](r)
		romFile = util.Must1(os.Open(path))
		r.variables["rom_file"] = romFile
		r.variables["nds"] = util.Must1(nds.OpenROM(romFile))
		fmt.Println("rom opened")
		return nil
	})

	// Opens a file from the loaded ROM's filesystem
	// - string				fname
	// + []byte
	r.AddFunc("nds_fread", func(r *Runner) any {
		fname := ReadArg[string](r)

		rom := ReadVar[*nds.Rom](r, "nds")
		datFile := util.Must1(rom.Filesystem.Open(fname))
		content := util.Must1(io.ReadAll(datFile))
		datFile.Close()

		return content
	})

	// Explain whats here at given ROM address
	// - uint32				address
	// + string
	r.AddFunc("here", func(r *Runner) any {
		addr := ReadArg[uint32](r)
		rom := ReadVar[*nds.Rom](r, "nds")
		entry := rom.WhatsHere(addr)
		if entry == nil {
			return "address not mapped"
		} else {
			return fmt.Sprintf(
				"%s [0x%X] [0x%X - 0x%X]\n",
				entry.Name(),
				addr-entry.From(),
				entry.From(),
				entry.From()+entry.To(),
			)
		}
	})

	// Opens a .dat entry from memory
	// - []byte				content
	// - int				index
	// + []byte
	r.AddFunc("dat_raw", func(r *Runner) any {
		content := ReadArg[[]byte](r)
		index := ReadArg[int](r)
		return pit.UnpackDatSingle(content, index)
	})

	// Opens a .dat file
	// - []byte				content
	r.AddFunc("dat_open", func(r *Runner) any {
		content := ReadArg[[]byte](r)
		r.variables["dat_open"] = pit.UnpackDat(content)
		fmt.Println("dat opened")
		return nil
	})

	// Opens a .dat entry
	// - int				index
	r.AddFunc("dat", func(r *Runner) any {
		index := ReadArg[int](r)
		dat := ReadVar[[][]byte](r, "dat_open")
		return dat[index]
	})

	// Get bitmap image from data
	// - int 				width
	// - int				height
	// - int				bit-depth (4 or 8)
	// - []byte				bitmap
	// - []byte				palette
	// + image.Image
	r.AddFunc("img_bitmap", func(r *Runner) any {
		w := ReadArg[int](r)
		h := ReadArg[int](r)
		bpp := ReadArg[int](r)
		gfx := ReadArg[[]byte](r)
		pal := nds.DeserializePalette(ReadArg[[]byte](r), true)
		return drawImgBitmap(w, h, bpp == 4, bytes.NewReader(gfx), pal)
	})

	// Get tilemap image from data
	// - int				width
	// - int				height
	// - int				bit-depth (4 or 8)
	// - []byte				tileset
	// - []byte				tilemap
	// - []byte				palette
	// + image.Image
	r.AddFunc("img_tilemap", func(r *Runner) any {
		w := ReadArg[int](r)
		h := ReadArg[int](r)
		bpp := ReadArg[int](r)
		gfx := ReadArg[[]byte](r)
		tlm := bytes.NewReader(ReadArg[[]byte](r))
		pal := nds.DeserializePalette(ReadArg[[]byte](r), true)

		var tls []nds.Tile
		if bpp == 4 {
			tls = nds.DeserializeTiles4BPP(gfx)
		} else if bpp == 8 {
			tls = nds.DeserializeTiles8BPP(gfx)
		} else {
			panic("unknown bitdepth for timemap, only 4 and 8 supported")
		}

		return drawImgTilemap(w, h, tls, tlm, pal)
	})

	// Get tilemap image from data
	// - int				width
	// - int				height
	// - int				bit-depth (4 or 8)
	// - []byte				tileset
	// - []byte				palette
	// + image.Image
	r.AddFunc("img_tileset", func(r *Runner) any {
		w := ReadArg[int](r)
		h := ReadArg[int](r)
		bpp := ReadArg[int](r)
		gfx := ReadArg[[]byte](r)
		pal := nds.DeserializePalette(ReadArg[[]byte](r), true)

		var tls []nds.Tile
		if bpp == 4 {
			tls = nds.DeserializeTiles4BPP(gfx)
		} else if bpp == 8 {
			tls = nds.DeserializeTiles8BPP(gfx)
		} else {
			panic("unknown bitdepth for timemap, only 4 and 8 supported")
		}

		// Build tilemap buffer
		tlmbuf := make([]byte, len(tls)*2)
		for i := range tls {
			binary.LittleEndian.PutUint16(tlmbuf[i*2:], uint16(i))
		}

		// Yup, do thing
		tlm := bytes.NewReader(tlmbuf)
		return drawImgTilemap(w, h, tls, tlm, pal)
	})

	// Exports a palette
	// - []byte				palette
	// + image.Image
	r.AddFunc("pal_export", func(r *Runner) any {
		pal := nds.DeserializePalette(ReadArg[[]byte](r), false)
		img := image.NewPaletted(image.Rect(0, 0, 256, 16*len(pal)/16), pal)

		for i := range pal {
			x := (i & 15) * 16
			y := (i >> 4) * 16

			for j := 0; j < 16; j++ {
				for k := 0; k < 16; k++ {
					img.SetColorIndex(x+j, y+k, uint8(i))
				}
			}
		}

		return img
	})
}

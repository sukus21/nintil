package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"slices"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/nds/pit"
	"github.com/sukus21/nintil/util"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage:", os.Args[0], "<PiT ROM path>")
		return
	}

	// Open ROM file
	romFile := util.Must1(os.Open(os.Args[1]))
	rom := util.Must1(nds.OpenROM(romFile))

	// Find FMapInfo
	fmapInfo := findFmapInfo(rom)
	fmapFile := util.Must1(rom.Filesystem.Open("FMap/FMapData.dat"))
	fmapReader := util.Must1(pit.NewFMapReader(fmapInfo, bytes.NewReader(util.Must1(io.ReadAll(fmapFile)))))

	for i := range fmapReader.MapCount() {
		os.RemoveAll(fmt.Sprintf("FMap_%d.png", i))
	}

	for i := range fmapReader.MapCount() {
		if err := RenderMap(fmapReader, i); err != nil {
			fmt.Println(err)
		}
	}
}

func findFmapInfo(rom *nds.Rom) []byte {
	size := 638 * 20
	for i := range rom.Arm9Binary[:len(rom.Arm9Binary)-size] {
		if slices.Equal([]byte{0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0}, rom.Arm9Binary[i:i+16]) {
			return rom.Arm9Binary[i : i+size]
		}
	}

	panic("could not find the thing :(")
}

func RenderMap(fmapReader *pit.FMapReader, mapId int) error {
	// Create renderer
	renderer, err := fmapReader.NewRenderer(mapId)
	if err != nil {
		return err
	}

	// Render the full map
	img, err := renderer.RenderMap()
	if err != nil {
		return err
	}
	err = exportImage(img, fmt.Sprintf("FMap_%d.png", mapId))
	if err != nil {
		return err
	}

	// Render layers
	for i := range 3 {
		if renderer.HasLayer(i) {
			img, err := renderer.RenderLayer(i)
			if err != nil {
				return err
			}
			err = exportImage(img, fmt.Sprintf("FMap_%d_%d.png", mapId, i))
			if err != nil {
				return err
			}
		}
	}

	// All good
	return nil
}

func exportImage(img image.Image, fname string) (err error) {
	of, err := os.Create(fname)
	if err != nil {
		return err
	}

	if err := png.Encode(of, img); err != nil {
		return err
	}

	return of.Close()
}

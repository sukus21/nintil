package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/nds/pit"
	"github.com/sukus21/nintil/util"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage:", os.Args[0], "<PiT ROM path>")
		return
	}

	// Open ROM file and create reader
	romFile := util.Must1(os.Open(os.Args[1]))
	rom := util.Must1(nds.OpenROM(romFile))
	fmapReader := util.Must1(pit.NewFMapReaderFromRom(rom))

	// Remove previous export
	os.RemoveAll("_FMap")
	os.Mkdir("_FMap", os.ModePerm)

	// Export all maps
	for i := range fmapReader.MapCount() {
		fmt.Println("exporting FMap", i, "...")
		if err := RenderMap(fmapReader, i); err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("done!")
}

func RenderMap(fmapReader *pit.FMapReader, mapId int) error {
	// Create fmap
	fmap, err := fmapReader.OpenMap(mapId)
	if err != nil {
		return err
	}

	// Render the full map
	img, err := fmap.RenderMap()
	if err != nil {
		return err
	}
	if err := exportImage(img, fmt.Sprintf("_FMap/FMap_%d.png", mapId)); err != nil {
		return err
	}
	if err := os.Mkdir(fmt.Sprintf("_FMap/FMap_%d", mapId), os.ModePerm); err != nil {
		return err
	}

	// Render layers
	for i := range 3 {
		if fmap.HasLayer(i) {
			img, err := fmap.RenderLayer(i)
			if err != nil {
				return err
			}
			err = exportImage(img, fmt.Sprintf("_FMap/FMap_%d/layer_%d.png", mapId, i))
			if err != nil {
				return err
			}
		}
	}

	// Dump metadata
	metadata, _ := json.MarshalIndent(fmap.Bundle.Metadata, "", "    ")
	os.WriteFile(fmt.Sprintf("_FMap/FMap_%d/metadata.json", mapId), metadata, os.ModePerm)

	// Dump treasure info
	if len(fmap.Treasure) != 0 {
		treasure, _ := json.MarshalIndent(fmap.Treasure, "", "    ")
		os.WriteFile(fmt.Sprintf("_FMap/FMap_%d/treasure.json", mapId), treasure, os.ModePerm)
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

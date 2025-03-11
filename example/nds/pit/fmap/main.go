package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/sukus21/nintil/nds/pit"
	"github.com/sukus21/nintil/util"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage:", os.Args[0], "<FMapInfo block file path>", "<FMapData.dat path>")
		return
	}

	fmapInfo := util.Must1(os.ReadFile(os.Args[1]))
	fmapFile := util.Must1(os.Open(os.Args[2]))
	fmapReader := util.Must1(pit.NewFMapReader(fmapInfo, fmapFile))

	for i := range fmapReader.MapCount() {
		os.RemoveAll(fmt.Sprintf("FMap_%d.png", i))
	}

	for i := range fmapReader.MapCount() {
		if err := RenderMap(fmapReader, i); err != nil {
			fmt.Println(err)
		}
	}
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

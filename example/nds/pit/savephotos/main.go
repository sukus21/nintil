package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/nds/pit"
	"github.com/sukus21/nintil/util"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage:", os.Args[0], "<PiT ROM path>")
		return
	}

	f := util.Must1(os.Open(os.Args[1]))
	defer f.Close()
	rom := util.Must1(nds.OpenROM(f))

	util.Must(os.MkdirAll("savephoto", os.ModePerm))
	for i := 0; i < 10; i++ {
		img := pit.DecodeSavePhoto(rom, i)
		ipath := fmt.Sprintf("savephotos/%d.png", i)
		outf := util.Must1(os.Create(ipath))
		png.Encode(outf, img)
		outf.Close()
	}
}

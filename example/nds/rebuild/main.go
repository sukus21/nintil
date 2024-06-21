package main

import (
	"log"
	"os"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("must specify ROM as command line parameter")
	}

	// Open ROM file
	in := util.Must1(os.Open(os.Args[1]))
	defer in.Close()
	rom := util.Must1(nds.OpenROM(in))

	// Rebuild ROM file
	out := util.Must1(os.Create("out.nds"))
	defer out.Close()
	util.Must(nds.SaveROM(rom, out))
}

package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sukus21/nintil/compression/lz10"
	"github.com/sukus21/nintil/compression/rlz"
	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/nds/pit"
	"github.com/sukus21/nintil/util"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("must specify ROM as command line parameter")
	}

	f := util.Must1(os.Open(os.Args[1]))
	defer f.Close()
	rom := util.Must1(nds.OpenROM(f))

	// Dump ARM9 binary in the stupidest way probably
	fout := util.Must1(os.Create("arm9.s"))
	defer fout.Close()
	for i := 0; i < len(rom.Arm9Binary); i += 4 {
		fmt.Fprintf(fout, "    .word 0x%08X\n", binary.LittleEndian.Uint32(rom.Arm9Binary[i:]))
	}
}

func writeFile(fname string, content []byte) error {
	if pit.IsDat(content) {
		util.Must(os.MkdirAll(fname, os.ModePerm))
		for i, content := range pit.UnpackDat(content) {
			outPath := filepath.Join(fname, strconv.Itoa(i))
			util.Must(writeFile(outPath, content))
		}
		return nil
	}

	if decomp, err := rlz.Decompress(content); err == nil {
		return os.WriteFile(fname+".rlz", decomp, os.ModePerm)
	}
	if decomp, err := lz10.Decompress(content); err == nil {
		return os.WriteFile(fname+".lz10", decomp, os.ModePerm)
	}

	return os.WriteFile(fname, content, os.ModePerm)
}

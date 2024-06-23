package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

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

	// Dump NitroFS filesystem
	util.Must(fs.WalkDir(rom.Filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		// Create directory for file
		outPath := filepath.Join("nitrofs", path)
		dname := filepath.Dir(outPath)
		util.Must(os.MkdirAll(dname, os.ModePerm))

		// Read file contents
		nitroFile := util.Must1(rom.Filesystem.Open(path))
		fileContent := util.Must1(io.ReadAll(nitroFile))
		nitroFile.Close()

		// Just dump regular files
		if filepath.Ext(path) != ".dat" || !pit.IsDat(fileContent) {
			util.Must(os.WriteFile(outPath, fileContent, os.ModePerm))
			return nil
		}

		// .dat file, NOW we are cooking!
		datPath := outPath
		util.Must(os.Mkdir(datPath, os.ModePerm))
		for i, content := range pit.UnpackDat(fileContent) {
			outPath := filepath.Join(datPath, fmt.Sprintf("%d", i))
			util.Must(os.WriteFile(outPath, content, os.ModePerm))
		}

		// Yup, all good
		return nil
	}))
}

// This example picks apart a ROM file into its components.
// The binaries are dumped, the file system exported, icon converted,
// the header exported as JSON, etc.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("must specify ROM as command line parameter")
	}

	f := util.Must1(os.Open(os.Args[1]))
	defer f.Close()
	rom := util.Must1(nds.OpenROM(f))

	// Dump header to JSON
	header := util.Must1(json.MarshalIndent(rom.GetHeader(), "", "\t"))
	util.Must(os.WriteFile("header.json", header, os.ModePerm))

	// Dump icon
	buf := &bytes.Buffer{}
	util.Must(png.Encode(buf, rom.GetIcon()))
	util.Must(os.WriteFile("icon.png", buf.Bytes(), os.ModePerm))

	// Dump title(s)
	titles := map[string]string{}
	for i := nds.TitleLanguage(0); nds.TitleLanguage(i) < nds.TitleLanguage_Count; i++ {
		title, err := rom.GetTitle(i)
		if err != nil {
			break
		}
		titles[i.String()] = title
	}
	titleJson := util.Must1(json.MarshalIndent(titles, "", "\t"))
	util.Must(os.WriteFile("title.json", titleJson, os.ModePerm))

	// Dump ARM7/ARM9 binaries
	util.Must(os.WriteFile("ARM9.bin", rom.Arm9Binary, os.ModePerm))
	util.Must(os.WriteFile("ARM7.bin", rom.Arm7Binary, os.ModePerm))

	// Dump NitroFS filesystem
	util.Must(fs.WalkDir(rom.Filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		outPath := filepath.Join("nitrofs", path)
		dname := filepath.Dir(outPath)
		util.Must(os.MkdirAll(dname, os.ModePerm))
		f := util.Must1(os.Create(outPath))
		defer f.Close()

		nitroFile := util.Must1(rom.Filesystem.Open(path))
		defer nitroFile.Close()
		util.Must1(io.Copy(f, nitroFile))

		return nil
	}))

	// Dump overlays
	util.Must(os.MkdirAll("overlays9", os.ModePerm))
	for i, v := range rom.Filesystem.GetArm9Overlays() {
		util.Must(os.WriteFile(fmt.Sprintf("overlays9/%d", i), v.Data(), os.ModePerm))
	}
	util.Must(os.MkdirAll("overlays7", os.ModePerm))
	for i, v := range rom.Filesystem.GetArm7Overlays() {
		util.Must(os.WriteFile(fmt.Sprintf("overlays7/%d", i), v.Data(), os.ModePerm))
	}

	// Print mapping
	mapFile := util.Must1(os.Create("mapping.txt"))
	defer mapFile.Close()
	fmt.Fprint(mapFile, rom.String())
}

package main

import (
	"bytes"
	"fmt"
	"io"

	_ "embed"

	"github.com/sukus21/nintil/nds/nsb"
	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

//go:embed "K_Dragon.nsbmd"
var modelData []byte

func main() {
	modelBuffer := bytes.NewReader(modelData)
	modelStruct := util.Must1(ezbin.Decode[nsb.Container](modelBuffer))

	for i := range modelStruct.Subfiles {
		subfile := &modelStruct.Subfiles[i]

		switch subfile.Stamp {
		case "MDL0":
			modelBuffer.Seek(subfile.Offset, io.SeekStart)
			subfileMDL := util.Must1(ezbin.Decode[nsb.MDL](modelBuffer))
			_ = subfileMDL
			fmt.Println("bunger")
			/*
				models := make(map[string]*nsb.Model)
				for i := range subfileMDL.Models.Count {
					name := subfileMDL.Models.Names[i]
					offset := subfileMDL.Models.Data[i] + uint32(subfileMDL.Offset)

					modelBuffer.Seek(int64(offset), io.SeekStart)
					model := util.Must1(ezbin.Decode[*nsb.Model](modelBuffer))
					models[name] = model
				}
			*/
		default:
			fmt.Println("unknown subfile type", subfile.Stamp)
		}
	}
}

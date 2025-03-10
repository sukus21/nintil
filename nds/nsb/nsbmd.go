package nsb

import (
	"fmt"
	"io"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

type Nsbmd struct {
	Container
	MDL
}

func ReadNsbmd(src io.ReadSeeker) (nsbmd *Nsbmd, err error) {
	defer util.Recover(&err)

	// Get container and reader
	r := newReader(src)
	container := util.Must1(ezbin.Decode[Container](r))
	r.b = container.ByteOrder

	// Read subfiles
	for _, v := range container.SubfileOffets {
		pos := container.StartPos + int64(v)
		util.Must1(src.Seek(pos, io.SeekStart))

		mdl := util.Must1(ezbin.Decode[MDL](r))
		models := make(map[string]Model)
		for name, offset := range mdl.Models.Map() {
			pos = mdl.StartPos + int64(offset)
			util.Must1(src.Seek(pos, io.SeekStart))
			model := util.Must1(ezbin.Decode[Model](r))

			// Read inverse bind matrices (it seems the easiest)
			invBindMatrices := make([]InvBindMatrix, model.InvBindMatricesOffset)

			models[name] = model
		}

		fmt.Println("done reading mdl")
	}

	return
}

type Matrix4x3[T ezbin.Number] [12]T

func (m *Matrix4x3[T]) Mat4x4() Matrix4x4[T] {
	return Matrix4x4[T]{
		m[0], m[1], m[2], 0,
		m[3], m[4], m[5], 0,
		m[6], m[7], m[8], 0,
		m[9], m[10], m[11], 1,
	}
}

type Matrix4x4[T ezbin.Number] [16]T
type InvBindMatrix struct {
	Matrix Matrix4x3[Fixed32]
	_      [9]Fixed32
}

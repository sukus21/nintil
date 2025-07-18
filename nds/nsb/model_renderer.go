package nsb

import (
	"bytes"
	"fmt"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
)

const (
	RENCMD_NOP_00       = 0x00
	RENCMD_NOP_40       = 0x40
	RENCMD_NOP_80       = 0x80
	RENCMD_END          = 0x01
	RENCMD_UNKNOWN_02   = 0x02
	RENCMD_MTX_LOAD     = 0x03
	RENCMD_MAT_BIND_04  = 0x04
	RENCMD_MAT_BIND_24  = 0x24
	RENCMD_MAT_BIND_44  = 0x44
	RENCMD_DRAW_MESH    = 0x05
	RENCMD_MULT_BONE    = 0x06
	RENCMD_MULT_BONE_R  = 0x26
	RENCMD_MULT_BONE_W  = 0x46
	RENCMD_MULT_BONE_RW = 0x66
	RENCMD_UNKNOWN_07   = 0x07
	RENCMD_UNKNOWN_47   = 0x47
	RENCMD_UNKNOWN_08   = 0x08
	RENCMD_SKIN         = 0x09
	RENCMD_SCALE_UP     = 0x0B
	RENCMD_SCALE_DOWN   = 0x2B
	RENCMD_UNKNOWN_0C   = 0x0C
	RENCMD_UNKNOWN_0D   = 0x0D
)

type NSBMD struct {
	Models   map[string]*Model
	Textures *SubfileTEX0
}

func (nsbmd *NSBMD) RenderModel(gpu *nds.GPU, modelID int) {
	/*
		model := &nsbmd.Models.Data[modelID]

		// TODO: no???
		cmds := bytes.NewReader(model.RenderCmds)
		runCmds := true
		for runCmds {
			opcode := util.Must1(cmds.ReadByte())

			switch opcode {
			case RENCMD_NOP_00, RENCMD_NOP_40, RENCMD_NOP_80:
				// Do nothing (wow, that's what a no-op is)

			case RENCMD_END:
				runCmds = false

			case RENCMD_UNKNOWN_02:
				util.Must1(cmds.ReadByte())
				util.Must1(cmds.ReadByte())

			case RENCMD_MTX_LOAD:

			case RENCMD_SKIN:
				storeIdx := util.Must1(cmds.ReadByte())
				numTerms := util.Must1(cmds.ReadByte())
				_ = storeIdx

				buf := make([]byte, numTerms)
				util.Must1(io.ReadFull(cmds, buf))

			default:
				fmt.Printf("TODO: render command %02X\n", opcode)
			}
		}

		// Render a single model instead
		renderer := &modelRenderer{
			gpu:   gpu,
			model: model,
		}

		renderer.RenderMesh(0)
	*/
}

type modelRenderer struct {
	gpu   *nds.GPU
	model *Model
}

func (ren *modelRenderer) RenderMesh(meshID int) {
	mesh := &ren.model.Meshes.Meshes.Data[meshID]

	cmds := bytes.NewReader(mesh.Commands)
	for cmds.Len() != 0 {
		packets := util.Must1(ren.gpu.ReadPackedCommands(cmds))

		for _, packet := range packets {
			util.Must(ren.gpu.ExecuteCommand(packet))
		}
	}

	fmt.Println("boo")
	_ = cmds
}

package g3d

import (
	"bytes"
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

type RenCMD uint8

const (
	RENCMD_NOP_00       = RenCMD(0x00)
	RENCMD_NOP_40       = RenCMD(0x40)
	RENCMD_NOP_80       = RenCMD(0x80)
	RENCMD_END          = RenCMD(0x01)
	RENCMD_UNKNOWN_02   = RenCMD(0x02)
	RENCMD_MTX_LOAD     = RenCMD(0x03)
	RENCMD_MAT_BIND_04  = RenCMD(0x04)
	RENCMD_MAT_BIND_24  = RenCMD(0x24)
	RENCMD_MAT_BIND_44  = RenCMD(0x44)
	RENCMD_DRAW_MESH    = RenCMD(0x05)
	RENCMD_MULT_BONE    = RenCMD(0x06)
	RENCMD_MULT_BONE_W  = RenCMD(0x26)
	RENCMD_MULT_BONE_R  = RenCMD(0x46)
	RENCMD_MULT_BONE_RW = RenCMD(0x66)
	RENCMD_UNKNOWN_07   = RenCMD(0x07)
	RENCMD_UNKNOWN_47   = RenCMD(0x47)
	RENCMD_UNKNOWN_08   = RenCMD(0x08)
	RENCMD_SKIN         = RenCMD(0x09)
	RENCMD_SCALE_UP     = RenCMD(0x0B)
	RENCMD_SCALE_DOWN   = RenCMD(0x2B)
	RENCMD_UNKNOWN_0C   = RenCMD(0x0C)
	RENCMD_UNKNOWN_0D   = RenCMD(0x0D)
)

type NSBMD struct {
	Models   map[string]*Model
	Textures *SubfileTEX0
}

func (nsbmd *NSBMD) RenderModel(gpu *nds.GPU, name string) {
	model := nsbmd.Models[name]

	// Render a single model instead
	renderer := &modelRenderer{
		gpu:   gpu,
		model: model,
		bones: make([]mgl64.Mat4, len(model.Bones)),
	}

	for i := range renderer.bones {
		renderer.bones[i] = model.Bones[i].Item.Matrix.Matrix
	}

	renderer.Render()
}

type modelRenderer struct {
	gpu   *nds.GPU
	model *Model
	bones []mgl64.Mat4
}

func matrixToParameters(mtx *mgl64.Mat4) []uint32 {
	out := make([]uint32, 16)
	for i, v := range mtx {
		out[i] = uint32(ezbin.EncodeFixedPoint(v, 1, 19, 12))
	}

	return out
}

func (ren *modelRenderer) Render() {
	for i := range ren.model.RenderCommands {
		cmd := &ren.model.RenderCommands[i]
		ren.ExecuteCommand(cmd)
	}
}

func (ren *modelRenderer) ExecuteCommand(cmd *RenderCommand) {
	switch cmd.Opcode {
	case RENCMD_NOP_00, RENCMD_NOP_40, RENCMD_NOP_80, RENCMD_END:
		// Do nothing (wow, that's what a no-op is)

	case RENCMD_UNKNOWN_02, RENCMD_UNKNOWN_07, RENCMD_UNKNOWN_08, RENCMD_UNKNOWN_0C, RENCMD_UNKNOWN_0D, RENCMD_UNKNOWN_47:
		// No clue what these do

	case RENCMD_MTX_LOAD:
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_RESTORE,
			Parameters: []uint32{uint32(cmd.Parameters[0])},
		})

	// Multiply bone thing
	case RENCMD_MULT_BONE:
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_MULT_4x4,
			Parameters: matrixToParameters(&ren.bones[cmd.Parameters[0]]),
		})
	case RENCMD_MULT_BONE_R:
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_RESTORE,
			Parameters: []uint32{uint32(cmd.Parameters[3])},
		})
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_MULT_4x4,
			Parameters: matrixToParameters(&ren.bones[cmd.Parameters[0]]),
		})
	case RENCMD_MULT_BONE_W:
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_MULT_4x4,
			Parameters: matrixToParameters(&ren.bones[cmd.Parameters[0]]),
		})
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_STORE,
			Parameters: []uint32{uint32(cmd.Parameters[3])},
		})
	case RENCMD_MULT_BONE_RW:
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_RESTORE,
			Parameters: []uint32{uint32(cmd.Parameters[4])},
		})
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_MULT_4x4,
			Parameters: matrixToParameters(&ren.bones[cmd.Parameters[0]]),
		})
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_STORE,
			Parameters: []uint32{uint32(cmd.Parameters[3])},
		})

	case RENCMD_SCALE_DOWN:
		scaleToFixed := uint32(ezbin.EncodeFixedPoint(ren.model.DownScale, 1, 19, 12))
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode: nds.GPUCMD_MTX_SCALE,
			Parameters: []uint32{
				scaleToFixed,
				scaleToFixed,
				scaleToFixed,
			},
		})

	case RENCMD_SCALE_UP:
		scaleToFixed := uint32(ezbin.EncodeFixedPoint(ren.model.UpScale, 1, 19, 12))
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode: nds.GPUCMD_MTX_SCALE,
			Parameters: []uint32{
				scaleToFixed,
				scaleToFixed,
				scaleToFixed,
			},
		})

	case RENCMD_SKIN:
		outputIdx := cmd.Parameters[0]
		numTerms := cmd.Parameters[1]

		mtx := mgl64.Mat4{}

		// Read terms
		for i := range numTerms {
			stackMatrixIdx := cmd.Parameters[2+i*3]
			invBindIdx := cmd.Parameters[3+i*3]
			weight := float64(cmd.Parameters[4+i*3]) / 256

			stackMatrix := &ren.gpu.MatrixStacks[1][stackMatrixIdx]
			invBindMatrix := &ren.model.Bones[invBindIdx].Item.InvBindMatrix.Matrix

			invBindMatrix4 := mgl64.Mat4FromRows(
				invBindMatrix.Row(0),
				invBindMatrix.Row(1),
				invBindMatrix.Row(2),
				mgl64.Vec4{0, 0, 0, 1},
			)

			mtx = mtx.Add(stackMatrix.Mul(weight).Mul4(invBindMatrix4))
		}

		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_LOAD_4x4,
			Parameters: matrixToParameters(&mtx),
		})
		ren.gpu.ExecuteCommand(nds.GPUCommand{
			Opcode:     nds.GPUCMD_MTX_STORE,
			Parameters: []uint32{uint32(outputIdx)},
		})

	case RENCMD_DRAW_MESH:
		ren.RenderMesh(int(cmd.Parameters[0]))

	default:
		fmt.Printf("TODO: render command 0x%02X %v\n", cmd.Opcode, cmd.Parameters)
	}
}

func (ren *modelRenderer) RenderMesh(meshID int) {
	mesh := &ren.model.Meshes[meshID].Item

	cmds := bytes.NewReader(mesh.Commands)
	for cmds.Len() != 0 {
		packets := util.Must1(ren.gpu.ReadPackedCommands(cmds))

		for _, packet := range packets {
			util.Must(ren.gpu.ExecuteCommand(packet))
		}
	}
}

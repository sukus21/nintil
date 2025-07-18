package nsb

import (
	"fmt"
	"io"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

func ReadNSBMD(r io.ReadSeeker) (_ *NSBMD, err error) {
	defer util.Recover(&err)

	// Read container
	container := util.Must1(ezbin.Decode[*Container](r))
	if container.Stamp != "BMD0" {
		return nil, fmt.Errorf("not a NSBMD file")
	}

	// Go thru subfiles
	out := new(NSBMD)
	for _, subfile := range container.Subfiles {
		switch subfile := subfile.Subfile.(type) {
		case *SubfileMDL0:
			out.Models = subfile.Models.Map()

		case *SubfileTEX0:
			out.Textures = subfile

		default:
			// return nil, fmt.Errorf("invalid subfile in NSBMD container")
		}
	}

	return out, nil
}

type SubfileMDL0 struct {
	_        struct{} `ezbin_tell:"namelist_offset"`
	Stamp    string   `ezbin_string:"ascii,4"`
	Filesize uint32
	Models   NameListOffset[*Model]
}

func (s *SubfileMDL0) GetStamp() string {
	return "MDL0"
}

type Model struct {
	Offset   int64 `ezbin_tell:"model"`
	FileSize uint32

	RenderCmds            RenderCommandList `ezbin_offset:"u32,model"`
	Materials             MaterialList      `ezbin_offset:"u32,model"`
	Meshes                MeshList          `ezbin_offset:"u32,model"`
	InvBindMatricesOffset uint32

	_               [3]byte
	NumBoneMatrices uint8
	NumMaterials    uint8
	NumMeshes       uint8
	_               [2]byte

	UpScale   float64 `ezbin_fixedpoint:"1,19,12"`
	DownScale float64 `ezbin_fixedpoint:"1,19,12"`

	NumVerts uint16
	NumPolys uint16
	NumTris  uint16
	NumQuads uint16

	BoundingBox BoundingBox
	_           [8]uint8
	BoneList    BoneList
}

type BoundingBox struct {
	MinX float64 `ezbin_fixedpoint:"1,3,12"`
	MinY float64 `ezbin_fixedpoint:"1,3,12"`
	MinZ float64 `ezbin_fixedpoint:"1,3,12"`
	MaxX float64 `ezbin_fixedpoint:"1,3,12"`
	MaxY float64 `ezbin_fixedpoint:"1,3,12"`
	MaxZ float64 `ezbin_fixedpoint:"1,3,12"`
}

type RenderCommandList struct {
	Commands []RenderCommand
}

func (l *RenderCommandList) EzbinDecode(r io.Reader) (err error) {
	defer util.Recover(&err)
	l.Commands = make([]RenderCommand, 0, 128)

	for {
		command := RenderCommand{}
		command.Opcode = ezbin.ReadSingle[uint8](r)

		switch command.Opcode {
		case RENCMD_END:
			return nil

		// No parameters
		case RENCMD_NOP_00, RENCMD_NOP_40, RENCMD_NOP_80,
			RENCMD_SCALE_DOWN, RENCMD_SCALE_UP:

		// 1 parameter
		case RENCMD_MTX_LOAD,
			RENCMD_MAT_BIND_04, RENCMD_MAT_BIND_24, RENCMD_MAT_BIND_44,
			RENCMD_DRAW_MESH,
			RENCMD_UNKNOWN_07, RENCMD_UNKNOWN_47,
			RENCMD_UNKNOWN_08:
			command.Parameters = make([]uint8, 1)
			util.Must1(io.ReadFull(r, command.Parameters))

		// 2 parameters
		case RENCMD_UNKNOWN_02,
			RENCMD_UNKNOWN_0C,
			RENCMD_UNKNOWN_0D:
			command.Parameters = make([]uint8, 2)
			util.Must1(io.ReadFull(r, command.Parameters))

		// 3 parameters
		case RENCMD_MULT_BONE:
			command.Parameters = make([]uint8, 3)
			util.Must1(io.ReadFull(r, command.Parameters))

		// 4 parameters
		case RENCMD_MULT_BONE_R, RENCMD_MULT_BONE_W:
			command.Parameters = make([]uint8, 4)
			util.Must1(io.ReadFull(r, command.Parameters))

		// 5 parameters
		case RENCMD_MULT_BONE_RW:
			command.Parameters = make([]uint8, 5)
			util.Must1(io.ReadFull(r, command.Parameters))

		// This one, variable parameters
		case RENCMD_SKIN:
			storeIndex := ezbin.ReadSingle[uint8](r)
			numTerms := ezbin.ReadSingle[uint8](r)

			command.Parameters = make([]uint8, 2+int(numTerms)*3)
			command.Parameters[0] = storeIndex
			command.Parameters[1] = numTerms
			util.Must1(io.ReadFull(r, command.Parameters[2:]))

		// What?
		default:
			return fmt.Errorf("invalid render command 0x%02X", command.Opcode)
		}

		l.Commands = append(l.Commands, command)
	}
}

type RenderCommand struct {
	Opcode     uint8
	Parameters []uint8
}

type MeshList struct {
	Offset int64 `ezbin_tell:"namelist_offset"`
	Meshes NameListOffset[Mesh]
}

type Mesh struct {
	Offset         int64 `ezbin_tell:"mesh"`
	_              uint16
	Size           uint16
	_              uint32
	CommandsOffset uint32
	CommandsLen    uint32
	Commands       []byte `ezbin_length:"CommandsLen" ezbin_offset:"CommandsOffset,mesh"`
}

type MaterialList struct {
	Offset                int64 `ezbin_tell:""`
	TexturePairingsOffset uint32
	PalettePairingsOffset uint32
	Materials             NameListOffset[uint32]
}

type BoneList struct {
	Offset int64 `ezbin_tell:"namelist_offset"`
	Bones  NameListOffset[BoneMatrix]
}

type boneMatrixFlags uint16

func (f boneMatrixFlags) hasTranslation() bool { return !ezbin.BitgetFlag(f, 1) }
func (f boneMatrixFlags) hasRotation() bool    { return !ezbin.BitgetFlag(f, 2) }
func (f boneMatrixFlags) hasScale() bool       { return !ezbin.BitgetFlag(f, 3) }
func (f boneMatrixFlags) hasPivot() bool       { return ezbin.BitgetFlag(f, 4) }

type Vector3 struct {
	X, Y, Z float64 `ezbin_fixedpoint:"1,19,12"`
}

type BoneRotationMatrix struct {
	M [8]float64 `ezbin_fixedpoint:"1,3,12"`
}

type BoneMatrix struct {
	Offset      int64 `ezbin_tell:""`
	Flags       boneMatrixFlags
	M0          float64            `ezbin_fixedpoint:"1,3,12"`
	Translation Vector3            `ezbin_ignore:""`
	Rotation    BoneRotationMatrix `ezbin_ignore:""`
	Scale       Vector3            `ezbin_ignore:""`
}

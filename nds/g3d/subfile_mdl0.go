package g3d

import (
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl64"
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
	RenderCommands []RenderCommand
	Meshes         []NamedItem[Mesh]
	Bones          []NamedItem[Bone]

	UpScale   float64
	DownScale float64

	NumVerts int
	NumPolys int
	NumTris  int
	NumQuads int

	BoundingBox BoundingBox
}

func (model *Model) EzbinDecode(r io.Reader) (err error) {
	defer util.Recover(&err)

	rs, ok := r.(io.ReadSeeker)
	if !ok {
		return ezbin.ErrNotSeeker
	}

	raw := util.Must1(ezbin.Decode[struct {
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
	}](rs))

	oldOffset := util.Must1(ezbin.At[int64](rs))
	util.Must1(rs.Seek(raw.Offset+int64(raw.InvBindMatricesOffset), io.SeekStart))

	invBindMatrices := make([]InvBindMatrix, raw.NumBoneMatrices)
	for i := range invBindMatrices {
		invBindMatrices[i] = util.Must1(ezbin.Decode[InvBindMatrix](rs))
	}

	// Begin transferring stuff to the model itself
	model.RenderCommands = raw.RenderCmds.Commands
	model.UpScale = raw.UpScale
	model.DownScale = raw.DownScale
	model.NumVerts = int(raw.NumVerts)
	model.NumPolys = int(raw.NumPolys)
	model.NumTris = int(raw.NumTris)
	model.NumQuads = int(raw.NumQuads)
	model.BoundingBox = raw.BoundingBox

	// Copy over meshes
	model.Meshes = raw.Meshes.Meshes.Items()

	// Copy over bones
	model.Bones = make([]NamedItem[Bone], len(raw.BoneList.Bones.Data))
	for i := range int(raw.BoneList.Bones.Count) {
		model.Bones[i].Name = raw.BoneList.Bones.Names[i]
		model.Bones[i].Item.Matrix = raw.BoneList.Bones.Data[i]

		if i < len(invBindMatrices) {
			model.Bones[i].Item.InvBindMatrix = invBindMatrices[i]
		}
	}

	util.Must1(rs.Seek(oldOffset, io.SeekStart))
	return nil
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
		command.Opcode = ezbin.ReadSingle[RenCMD](r)

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
	Opcode     RenCMD
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

type Bone struct {
	Matrix        BoneMatrix
	InvBindMatrix InvBindMatrix
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

type BoneMatrix struct {
	Translation mgl64.Vec3
	Rotation    mgl64.Mat3
	Scale       mgl64.Vec3
	Matrix      mgl64.Mat4
}

func (bm *BoneMatrix) EzbinDecode(r io.Reader) (err error) {
	defer util.Recover(&err)

	f := ezbin.ReadSingle[boneMatrixFlags](r)
	m0 := util.Must1(ezbin.ReadFixedPoint(r, 1, 3, 12))

	// Translation
	bm.Translation = mgl64.Vec3{}
	if f.hasTranslation() {
		for i := range 3 {
			bm.Translation[i] = util.Must1(ezbin.ReadFixedPoint(r, 1, 19, 12))
		}
	}

	// Rotation
	bm.Rotation = mgl64.Ident3()
	if f.hasPivotMatrix() {
		a := util.Must1(ezbin.ReadFixedPoint(r, 1, 3, 12))
		b := util.Must1(ezbin.ReadFixedPoint(r, 1, 3, 12))
		bm.Rotation = getPivotMatrix(f, a, b)
	} else if f.hasRotation() {
		bm.Rotation[0] = m0
		for i := range 8 {
			bm.Rotation[i+1] = util.Must1(ezbin.ReadFixedPoint(r, 1, 3, 12))
		}
	}

	// Scale
	bm.Scale = mgl64.Vec3{1, 1, 1}
	if f.hasScale() {
		for i := range 3 {
			bm.Scale[i] = util.Must1(ezbin.ReadFixedPoint(r, 1, 19, 12))
		}
	}

	// Build matrix
	bm.Matrix = mgl64.Ident4()
	if f.hasScale() {
		scaleMtx := mgl64.Scale3D(bm.Scale.Elem())
		bm.Matrix = bm.Matrix.Mul4(scaleMtx)
	}
	if f.hasRotation() || f.hasPivotMatrix() {
		rotationMtx := bm.Rotation.Mat4()
		bm.Matrix = rotationMtx.Mul4(bm.Matrix)
	}
	if f.hasTranslation() {
		translationMtx := mgl64.Translate3D(bm.Translation.Elem())
		bm.Matrix = translationMtx.Mul4(bm.Matrix)
	}

	// Alrighty, that should be all
	return nil
}

type boneMatrixFlags uint16

func (f boneMatrixFlags) hasTranslation() bool { return !ezbin.BitgetFlag(f, 0) }
func (f boneMatrixFlags) hasRotation() bool    { return !ezbin.BitgetFlag(f, 1) }
func (f boneMatrixFlags) hasScale() bool       { return !ezbin.BitgetFlag(f, 2) }
func (f boneMatrixFlags) hasPivotMatrix() bool { return ezbin.BitgetFlag(f, 3) }
func (f boneMatrixFlags) getForm() int         { return ezbin.Bitget[int](f, 4, 4) }
func (f boneMatrixFlags) getNegOne() bool      { return ezbin.BitgetFlag(f, 8) }
func (f boneMatrixFlags) getNegC() bool        { return ezbin.BitgetFlag(f, 9) }
func (f boneMatrixFlags) getNegD() bool        { return ezbin.BitgetFlag(f, 10) }

type InvBindMatrix struct {
	Matrix  mgl64.Mat3x4 `ezbin_fixedpoint:"1,19,12"`
	Unknown mgl64.Mat3   `ezbin_fixedpoint:"1,19,12"`
}

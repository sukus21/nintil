package nds

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"io"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

type GPUOpcode uint8

const (
	GPUCMD_MTX_MODE       = GPUOpcode(0x10)
	GPUCMD_MTX_PUSH       = GPUOpcode(0x11)
	GPUCMD_MTX_POP        = GPUOpcode(0x12)
	GPUCMD_MTX_STORE      = GPUOpcode(0x13)
	GPUCMD_MTX_RESTORE    = GPUOpcode(0x14)
	GPUCMD_MTX_IDENTITY   = GPUOpcode(0x15)
	GPUCMD_MTX_LOAD_4x4   = GPUOpcode(0x16)
	GPUCMD_MTX_LOAD_4x3   = GPUOpcode(0x17)
	GPUCMD_MTX_MULT_4x4   = GPUOpcode(0x18)
	GPUCMD_MTX_MULT_4x3   = GPUOpcode(0x19)
	GPUCMD_MTX_MULT_3x3   = GPUOpcode(0x1A)
	GPUCMD_MTX_SCALE      = GPUOpcode(0x1B)
	GPUCMD_MTX_TRANS      = GPUOpcode(0x1C)
	GPUCMD_COLOR          = GPUOpcode(0x20)
	GPUCMD_NORMAL         = GPUOpcode(0x21)
	GPUCMD_TEXCOORD       = GPUOpcode(0x22)
	GPUCMD_VTX_16         = GPUOpcode(0x23)
	GPUCMD_VTX_10         = GPUOpcode(0x24)
	GPUCMD_VTX_XY         = GPUOpcode(0x25)
	GPUCMD_VTX_XZ         = GPUOpcode(0x26)
	GPUCMD_VTX_YZ         = GPUOpcode(0x27)
	GPUCMD_VTX_DIFF       = GPUOpcode(0x28)
	GPUCMD_POLYGON_ATTR   = GPUOpcode(0x29)
	GPUCMD_TEXIMAGE_PARAM = GPUOpcode(0x2A)
	GPUCMD_PLTT_BASE      = GPUOpcode(0x2B)
	GPUCMD_DIF_AMB        = GPUOpcode(0x30)
	GPUCMD_SPE_EMI        = GPUOpcode(0x31)
	GPUCMD_LIGHT_VECTOR   = GPUOpcode(0x32)
	GPUCMD_LIGHT_COLOR    = GPUOpcode(0x33)
	GPUCMD_SHININESS      = GPUOpcode(0x34)
	GPUCMD_BEGIN_VTXS     = GPUOpcode(0x40)
	GPUCMD_END_VTXS       = GPUOpcode(0x41)
)

var GpuCmdParameterNum = [256]int{
	GPUCMD_MTX_MODE:       1,
	GPUCMD_MTX_IDENTITY:   0,
	GPUCMD_MTX_LOAD_4x4:   16,
	GPUCMD_MTX_LOAD_4x3:   12,
	GPUCMD_MTX_MULT_4x4:   16,
	GPUCMD_MTX_MULT_4x3:   12,
	GPUCMD_MTX_MULT_3x3:   9,
	GPUCMD_MTX_SCALE:      3,
	GPUCMD_MTX_TRANS:      3,
	GPUCMD_MTX_PUSH:       0,
	GPUCMD_MTX_POP:        1,
	GPUCMD_MTX_STORE:      1,
	GPUCMD_MTX_RESTORE:    1,
	GPUCMD_POLYGON_ATTR:   1,
	GPUCMD_COLOR:          1,
	GPUCMD_BEGIN_VTXS:     1,
	GPUCMD_END_VTXS:       0,
	GPUCMD_VTX_16:         2,
	GPUCMD_VTX_10:         1,
	GPUCMD_VTX_XY:         1,
	GPUCMD_VTX_XZ:         1,
	GPUCMD_VTX_YZ:         1,
	GPUCMD_VTX_DIFF:       1,
	GPUCMD_LIGHT_VECTOR:   1,
	GPUCMD_LIGHT_COLOR:    1,
	GPUCMD_DIF_AMB:        1,
	GPUCMD_SPE_EMI:        1,
	GPUCMD_SHININESS:      1,
	GPUCMD_NORMAL:         1,
	GPUCMD_TEXCOORD:       1,
	GPUCMD_TEXIMAGE_PARAM: 1,
	GPUCMD_PLTT_BASE:      1,
}

var GpuCmdString = [256]string{
	GPUCMD_MTX_MODE:       "MTX_MODE",
	GPUCMD_MTX_IDENTITY:   "MTX_IDENTITY",
	GPUCMD_MTX_LOAD_4x4:   "MTX_LOAD_4X4",
	GPUCMD_MTX_LOAD_4x3:   "MTX_LOAD_4X3",
	GPUCMD_MTX_MULT_4x4:   "MTX_MULT_4X4",
	GPUCMD_MTX_MULT_4x3:   "MTX_MULT_4X3",
	GPUCMD_MTX_MULT_3x3:   "MTX_MULT_3X3",
	GPUCMD_MTX_SCALE:      "MTX_SCALE",
	GPUCMD_MTX_TRANS:      "MTX_TRANS",
	GPUCMD_MTX_PUSH:       "MTX_PUSH",
	GPUCMD_MTX_POP:        "MTX_POP",
	GPUCMD_MTX_STORE:      "MTX_STORE",
	GPUCMD_MTX_RESTORE:    "MTX_RESTORE",
	GPUCMD_POLYGON_ATTR:   "POLYGON_ATTR",
	GPUCMD_COLOR:          "COLOR",
	GPUCMD_BEGIN_VTXS:     "BEGIN_VTXS",
	GPUCMD_END_VTXS:       "END_VTXS",
	GPUCMD_VTX_16:         "VTX_16",
	GPUCMD_VTX_10:         "VTX_10",
	GPUCMD_VTX_XY:         "VTX_XY",
	GPUCMD_VTX_XZ:         "VTX_XZ",
	GPUCMD_VTX_YZ:         "VTX_YZ",
	GPUCMD_VTX_DIFF:       "VTX_DIFF",
	GPUCMD_LIGHT_VECTOR:   "LIGHT_VECTOR",
	GPUCMD_LIGHT_COLOR:    "LIGHT_COLOR",
	GPUCMD_DIF_AMB:        "DIF_AMB",
	GPUCMD_SPE_EMI:        "SPE_EMI",
	GPUCMD_SHININESS:      "SHININESS",
	GPUCMD_NORMAL:         "NORMAL",
	GPUCMD_TEXCOORD:       "TEXCOORD",
	GPUCMD_TEXIMAGE_PARAM: "TEXIMAGE_PARAM",
	GPUCMD_PLTT_BASE:      "PLTT_BASE",
}

func gpuOpcodeToString(cmd GPUOpcode) string {
	str := GpuCmdString[cmd]
	if str == "" {
		return fmt.Sprintf("INVALID <0x%02X>", cmd)
	} else {
		return str
	}
}

type MatrixMode int

const (
	MatrixMode_Projection  = MatrixMode(0)
	MatrixMode_Coordinate  = MatrixMode(1)
	MatrixMode_Directional = MatrixMode(2)
	MatrixMode_Texture     = MatrixMode(3)
)

type PolygonMode int

const (
	PolygonMode_TriangleList  = PolygonMode(0)
	PolygonMode_QuadList      = PolygonMode(1)
	PolygonMode_TriangleStrip = PolygonMode(2)
	PolygonMode_QuadStrip     = PolygonMode(3)
)

type GPU struct {
	MatrixStacks [4][]mgl64.Mat4
	Matrices     [4]mgl64.Mat4
	StackPos     []int

	w GpuOutput

	// Internal state
	polyVerts   int
	matrixMode  MatrixMode
	polygonMode PolygonMode
	vertex      Vertex
	vertexCache [4]Vertex
	oddTriStrip bool

	// Stats
	numPolys int
	numTris  int
	numQuads int
}

type GPUCommand struct {
	Opcode     GPUOpcode
	Parameters []uint32
}

type Vertex struct {
	Position  mgl64.Vec3
	TexCoords mgl64.Vec2
	Normal    mgl64.Vec3
	Color     color.RGBA
}

type GpuOutput struct {
	// Written to whenever a vertex is submitted.
	VertWriter chan<- Vertex

	// Written to whenever a polygon (tri or quad) is submitted.
	// If a quad is submitted, 2 tris will be pushed here.
	// If a tri is submitted, only that tri is pushed.
	PolyWriter chan<- [3]Vertex

	// Written to whenever a tri is submitted.
	TriWriter chan<- [3]Vertex

	// Written to whenever a quad is submitted.
	QuadWriter chan<- [4]Vertex
}

func NewGPU(dst GpuOutput) *GPU {
	gpu := &GPU{
		w: dst,
		MatrixStacks: [4][]mgl64.Mat4{
			make([]mgl64.Mat4, 1),
			make([]mgl64.Mat4, 31),
			make([]mgl64.Mat4, 31),
			make([]mgl64.Mat4, 1),
		},
	}

	for i := range 4 {
		gpu.Matrices[i] = mgl64.Ident4()
		for j := range gpu.MatrixStacks[i] {
			gpu.MatrixStacks[i][j] = mgl64.Ident4()
		}
	}

	return gpu
}

func (gpu *GPU) ReadUnpackedCommand(r io.Reader) (_ GPUCommand, err error) {
	defer util.Recover(&err)

	opcode := GPUOpcode(ezbin.ReadSingle[uint32](r))
	parameters := make([]uint32, GpuCmdParameterNum[opcode])
	util.Must(binary.Read(r, binary.LittleEndian, parameters))

	return GPUCommand{
		Opcode:     opcode,
		Parameters: parameters,
	}, nil
}

func (gpu *GPU) ReadPackedCommands(r io.Reader) (_ [4]GPUCommand, err error) {
	defer util.Recover(&err)

	packedOpcodes := ezbin.ReadSingle[[4]GPUOpcode](r)
	out := [4]GPUCommand{}
	for i := range packedOpcodes {
		cmd := &out[i]

		cmd.Opcode = packedOpcodes[i]
		cmd.Parameters = make([]uint32, GpuCmdParameterNum[cmd.Opcode])
		util.Must(binary.Read(r, binary.LittleEndian, cmd.Parameters))
	}

	return out, nil
}

var AllowedPolys = 1

func (gpu *GPU) submitVertex() {

	// Transform matrix position coordinates
	vertex := gpu.vertex
	coordinateMatrix := &gpu.Matrices[MatrixMode_Coordinate]
	vertex.Position = coordinateMatrix.Mul4x1(vertex.Position.Vec4(1)).Vec3()

	// Add vertex to vertex cache
	gpu.vertexCache[gpu.polyVerts] = vertex
	gpu.polyVerts++
	if gpu.w.VertWriter != nil {
		gpu.w.VertWriter <- vertex
	}

	// Do we have a completed polygon?
	polygon, ok := gpu.getCompletePoly()
	if !ok {
		return
	}

	gpu.numPolys++
	if len(polygon) == 3 {
		gpu.numTris++

		if gpu.w.PolyWriter != nil || gpu.w.TriWriter != nil {
			tri := [3]Vertex{}
			copy(tri[:], polygon)

			if gpu.w.TriWriter != nil {
				gpu.w.TriWriter <- tri
			}
			if gpu.w.PolyWriter != nil {
				gpu.w.PolyWriter <- tri
			}
		}
	} else {
		gpu.numQuads++

		if gpu.w.PolyWriter != nil || gpu.w.QuadWriter != nil {
			quad := [4]Vertex{}
			copy(quad[:], polygon)

			if gpu.w.QuadWriter != nil {
				gpu.w.QuadWriter <- quad
			}
			if gpu.w.PolyWriter != nil {
				gpu.w.PolyWriter <- [3]Vertex{
					quad[0],
					quad[1],
					quad[2],
				}
				gpu.w.PolyWriter <- [3]Vertex{
					quad[0],
					quad[2],
					quad[3],
				}
			}
		}
	}
}

func (gpu *GPU) getCompletePoly() ([]Vertex, bool) {
	switch {
	case gpu.polygonMode == PolygonMode_TriangleList && gpu.polyVerts == 3:
		out := make([]Vertex, 3)
		copy(out, gpu.vertexCache[:])
		gpu.polyVerts = 0
		return out, true

	case gpu.polygonMode == PolygonMode_TriangleStrip && gpu.polyVerts == 3:
		out := make([]Vertex, 3)
		copy(out, gpu.vertexCache[:])
		gpu.vertexCache[0] = gpu.vertexCache[1]
		gpu.vertexCache[1] = gpu.vertexCache[2]
		gpu.polyVerts = 2

		// Face the correct direction
		if gpu.oddTriStrip {
			out[0], out[1] = out[1], out[0]
		}
		gpu.oddTriStrip = !gpu.oddTriStrip

		return out, true

	case gpu.polygonMode == PolygonMode_QuadList && gpu.polyVerts == 4:
		out := make([]Vertex, 4)
		copy(out, gpu.vertexCache[:])
		gpu.polyVerts = 0
		return out, true

	case gpu.polygonMode == PolygonMode_QuadStrip && gpu.polyVerts == 4:
		out := make([]Vertex, 4)
		copy(out, gpu.vertexCache[:])
		out[2], out[3] = out[3], out[2]
		gpu.polyVerts = 2
		return out, true
	}

	return nil, false
}

func (gpu *GPU) currentMatrix() *mgl64.Mat4 {
	return &gpu.Matrices[gpu.matrixMode]
}

func (gpu *GPU) currentMatrixStack() *[]mgl64.Mat4 {
	return &gpu.MatrixStacks[gpu.matrixMode]
}

func (gpu *GPU) ExecuteCommand(cmd GPUCommand) error {
	if len(cmd.Parameters) != GpuCmdParameterNum[cmd.Opcode] {
		return errors.New("nds.GPU: incorrect parameter count for opcode")
	}

	switch cmd.Opcode {
	case 0:
		// no-op

	case GPUCMD_BEGIN_VTXS:
		gpu.polygonMode = PolygonMode(cmd.Parameters[0] & 3)
		gpu.polyVerts = 0
		gpu.oddTriStrip = false

	case GPUCMD_END_VTXS:
		// no-op

	case GPUCMD_NORMAL:
		c := cmd.Parameters[0]
		gpu.vertex.Normal[0] = ezbin.DecodeFixedPoint(uint64(c>>0)&0x3F, 1, 0, 9)
		gpu.vertex.Normal[1] = ezbin.DecodeFixedPoint(uint64(c>>10)&0x3F, 1, 0, 9)
		gpu.vertex.Normal[2] = ezbin.DecodeFixedPoint(uint64(c>>20)&0x3F, 1, 0, 9)

	case GPUCMD_TEXCOORD:
		c := cmd.Parameters[0]
		gpu.vertex.TexCoords[0] = ezbin.DecodeFixedPoint(uint64(c&0xFFFF), 1, 11, 4)
		gpu.vertex.TexCoords[1] = ezbin.DecodeFixedPoint(uint64(c>>16), 1, 11, 4)

	case GPUCMD_COLOR:
		c := cmd.Parameters[0]
		gpu.vertex.Color.R = uint8(c>>0) & 0x1F
		gpu.vertex.Color.G = uint8(c>>5) & 0x1F
		gpu.vertex.Color.B = uint8(c>>10) & 0x1F
		gpu.vertex.Color.A = 255

	case GPUCMD_VTX_16:
		gpu.vertex.Position[0] = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]&0xFFFF), 1, 3, 12)
		gpu.vertex.Position[1] = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]>>16), 1, 3, 12)
		gpu.vertex.Position[2] = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[1]&0xFFFF), 1, 3, 12)
		gpu.submitVertex()

	case GPUCMD_VTX_XY:
		gpu.vertex.Position[0] = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]&0xFFFF), 1, 3, 12)
		gpu.vertex.Position[1] = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]>>16), 1, 3, 12)
		gpu.submitVertex()

	case GPUCMD_VTX_XZ:
		gpu.vertex.Position[0] = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]&0xFFFF), 1, 3, 12)
		gpu.vertex.Position[2] = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]>>16), 1, 3, 12)
		gpu.submitVertex()

	case GPUCMD_VTX_YZ:
		gpu.vertex.Position[1] = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]&0xFFFF), 1, 3, 12)
		gpu.vertex.Position[2] = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]>>16), 1, 3, 12)
		gpu.submitVertex()

	case GPUCMD_VTX_DIFF:
		scale := 1.0 / 8.0
		c := cmd.Parameters[0]
		gpu.vertex.Position[0] += ezbin.DecodeFixedPoint(uint64(c>>0)&0x3F, 1, 0, 9) * scale
		gpu.vertex.Position[1] += ezbin.DecodeFixedPoint(uint64(c>>10)&0x3F, 1, 0, 9) * scale
		gpu.vertex.Position[2] += ezbin.DecodeFixedPoint(uint64(c>>20)&0x3F, 1, 0, 9) * scale
		gpu.submitVertex()

	case GPUCMD_MTX_MODE:
		gpu.matrixMode = MatrixMode(cmd.Parameters[0])

	case GPUCMD_MTX_MULT_4x4:
		mtx := matrixFromParams(cmd.Parameters)
		*gpu.currentMatrix() = gpu.currentMatrix().Mul4(mtx)

	case GPUCMD_MTX_SCALE:
		vec := vectorFromParams(cmd.Parameters)
		mtx := mgl64.Scale3D(vec.Elem())
		*gpu.currentMatrix() = gpu.currentMatrix().Mul4(mtx)

	case GPUCMD_MTX_LOAD_4x4:
		mtx := matrixFromParams(cmd.Parameters)
		*gpu.currentMatrix() = mtx

	case GPUCMD_MTX_IDENTITY:
		*gpu.currentMatrix() = mgl64.Ident4()

	case GPUCMD_MTX_STORE:
		(*gpu.currentMatrixStack())[cmd.Parameters[0]] = *gpu.currentMatrix()

	case GPUCMD_MTX_RESTORE:
		*gpu.currentMatrix() = (*gpu.currentMatrixStack())[cmd.Parameters[0]]

	default:
		fmt.Printf("TODO: GPU command %s %v\n", gpuOpcodeToString(cmd.Opcode), cmd.Parameters)
	}

	return nil
}

func vectorFromParams(params []uint32) mgl64.Vec3 {
	out := mgl64.Vec3{}

	for i, v := range params {
		out[i] = ezbin.DecodeFixedPoint(uint64(v), 1, 19, 12)
	}

	return out
}

func matrixFromParams(params []uint32) mgl64.Mat4 {
	out := mgl64.Mat4{}

	for i, v := range params {
		out[i] = ezbin.DecodeFixedPoint(uint64(v), 1, 19, 12)
	}

	return out
}

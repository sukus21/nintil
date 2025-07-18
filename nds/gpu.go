package nds

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

const (
	GPUCMD_MTX_MODE       = 0x10
	GPUCMD_MTX_PUSH       = 0x11
	GPUCMD_MTX_POP        = 0x12
	GPUCMD_MTX_STORE      = 0x13
	GPUCMD_MTX_RESTORE    = 0x14
	GPUCMD_MTX_IDENTITY   = 0x15
	GPUCMD_MTX_LOAD_4x4   = 0x16
	GPUCMD_MTX_LOAD_4x3   = 0x17
	GPUCMD_MTX_MULT_4x4   = 0x18
	GPUCMD_MTX_MULT_4x3   = 0x19
	GPUCMD_MTX_MULT_3x3   = 0x1A
	GPUCMD_MTX_SCALE      = 0x1B
	GPUCMD_MTX_TRANS      = 0x1C
	GPUCMD_COLOR          = 0x20
	GPUCMD_NORMAL         = 0x21
	GPUCMD_TEXCOORD       = 0x22
	GPUCMD_VTX_16         = 0x23
	GPUCMD_VTX_10         = 0x24
	GPUCMD_VTX_XY         = 0x25
	GPUCMD_VTX_XZ         = 0x26
	GPUCMD_VTX_YZ         = 0x27
	GPUCMD_VTX_DIFF       = 0x28
	GPUCMD_POLYGON_ATTR   = 0x29
	GPUCMD_TEXIMAGE_PARAM = 0x2A
	GPUCMD_PLTT_BASE      = 0x2B
	GPUCMD_DIF_AMB        = 0x30
	GPUCMD_SPE_EMI        = 0x31
	GPUCMD_LIGHT_VECTOR   = 0x32
	GPUCMD_LIGHT_COLOR    = 0x33
	GPUCMD_SHININESS      = 0x34
	GPUCMD_BEGIN_VTXS     = 0x40
	GPUCMD_END_VTXS       = 0x41
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

func gpuCmdToString(cmd uint8) string {
	str := GpuCmdString[cmd]
	if str == "" {
		return fmt.Sprintf("INVALID <0x%02X>", cmd)
	} else {
		return str
	}
}

type gpuMatrix [16]float64

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
	ProjectionStack    [1]gpuMatrix
	ProjectionMatrix   gpuMatrix
	ProjectionStackPos int

	CoordinateStack    [31]gpuMatrix
	CoordinateMatrix   gpuMatrix
	CoordinateStackPos int

	DirectionalStack  [31]gpuMatrix
	DirectionalMatrix gpuMatrix

	w io.Writer

	// Internal state
	polyVerts   int
	matrixMode  MatrixMode
	polygonMode PolygonMode
	vertex      Vertex

	// Stats
	numPolys int
	numTris  int
	numQuads int

	vertexList []Vertex
}

type GPUCommand struct {
	Opcode     uint8
	Parameters []uint32
}

type Vertex struct {
	X, Y, Z    float64
	U, V       float64
	NX, NY, NZ float64
	R, G, B    uint8
}

func NewGPU(dst io.Writer) *GPU {
	return &GPU{
		w: dst,
	}
}

func (gpu *GPU) ReadUnpackedCommand(r io.Reader) (_ GPUCommand, err error) {
	defer util.Recover(&err)

	opcode := uint8(ezbin.ReadSingle[uint32](r))
	parameters := make([]uint32, GpuCmdParameterNum[opcode])
	util.Must(binary.Read(r, binary.LittleEndian, parameters))

	return GPUCommand{
		Opcode:     opcode,
		Parameters: parameters,
	}, nil
}

func (gpu *GPU) ReadPackedCommands(r io.Reader) (_ [4]GPUCommand, err error) {
	defer util.Recover(&err)

	packedOpcodes := ezbin.ReadSingle[[4]byte](r)
	out := [4]GPUCommand{}
	for i := range packedOpcodes {
		cmd := &out[i]

		cmd.Opcode = packedOpcodes[i]
		cmd.Parameters = make([]uint32, GpuCmdParameterNum[cmd.Opcode])
		util.Must(binary.Read(r, binary.LittleEndian, cmd.Parameters))
	}

	return out, nil
}

func (gpu *GPU) submitVertex() {
	gpu.vertexList = append(gpu.vertexList, gpu.vertex)
	gpu.polyVerts++

	if gpu.isPolyComplete() {
		gpu.numPolys++
		if gpu.polygonMode&1 == 0 {
			gpu.numTris++
		} else {
			gpu.numQuads++
		}
	}
}

func (gpu *GPU) isPolyComplete() bool {
	switch gpu.polygonMode {
	case PolygonMode_TriangleList:
		if gpu.polyVerts == 3 {
			gpu.polyVerts = 0
			return true
		}

	case PolygonMode_TriangleStrip:
		if gpu.polyVerts == 3 {
			gpu.polyVerts = 2
			return true
		}

	case PolygonMode_QuadList:
		if gpu.polyVerts == 4 {
			gpu.polyVerts = 0
			return true
		}

	case PolygonMode_QuadStrip:
		if gpu.polyVerts == 4 {
			gpu.polyVerts = 2
			return true
		}
	}

	return false
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

	case GPUCMD_END_VTXS:
		// no-op

	case GPUCMD_NORMAL:
		c := cmd.Parameters[0]
		gpu.vertex.NX = ezbin.DecodeFixedPoint(uint64(c>>0)&0x3F, 1, 0, 9)
		gpu.vertex.NY = ezbin.DecodeFixedPoint(uint64(c>>10)&0x3F, 1, 0, 9)
		gpu.vertex.NZ = ezbin.DecodeFixedPoint(uint64(c>>20)&0x3F, 1, 0, 9)

	case GPUCMD_TEXCOORD:
		c := cmd.Parameters[0]
		gpu.vertex.U = ezbin.DecodeFixedPoint(uint64(c&0xFFFF), 1, 11, 4)
		gpu.vertex.V = ezbin.DecodeFixedPoint(uint64(c>>16), 1, 11, 4)

	case GPUCMD_COLOR:
		c := cmd.Parameters[0]
		gpu.vertex.R = uint8(c>>0) & 0x1F
		gpu.vertex.R = uint8(c>>5) & 0x1F
		gpu.vertex.R = uint8(c>>10) & 0x1F

	case GPUCMD_VTX_16:
		gpu.vertex.X = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]&0xFFFF), 1, 3, 12)
		gpu.vertex.Y = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]>>16), 1, 3, 12)
		gpu.vertex.Z = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[1]&0xFFFF), 1, 3, 12)
		gpu.submitVertex()

	case GPUCMD_VTX_XY:
		gpu.vertex.X = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]&0xFFFF), 1, 3, 12)
		gpu.vertex.Y = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]>>16), 1, 3, 12)
		gpu.submitVertex()

	case GPUCMD_VTX_XZ:
		gpu.vertex.X = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]&0xFFFF), 1, 3, 12)
		gpu.vertex.Z = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]>>16), 1, 3, 12)
		gpu.submitVertex()

	case GPUCMD_VTX_YZ:
		gpu.vertex.Y = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]&0xFFFF), 1, 3, 12)
		gpu.vertex.Z = ezbin.DecodeFixedPoint(uint64(cmd.Parameters[0]>>16), 1, 3, 12)
		gpu.submitVertex()

	case GPUCMD_VTX_DIFF:
		scale := 1.0 / 8.0
		c := cmd.Parameters[0]
		gpu.vertex.X += ezbin.DecodeFixedPoint(uint64(c>>0)&0x3F, 1, 0, 9) * scale
		gpu.vertex.Y += ezbin.DecodeFixedPoint(uint64(c>>10)&0x3F, 1, 0, 9) * scale
		gpu.vertex.Z += ezbin.DecodeFixedPoint(uint64(c>>20)&0x3F, 1, 0, 9) * scale
		gpu.submitVertex()

	default:
		fmt.Println(gpuCmdToString(cmd.Opcode))
	}

	return nil
}

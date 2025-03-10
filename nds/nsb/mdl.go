package nsb

type MDL struct {
	StartPos int64  `ezbin_seekpos:""`
	Stamp    string `ezbin_string:"ascii" ezbin_length:"4"`
	Filesize uint32
	Models   Namelist[uint32]
}

type Model struct {
	StartPos              int64 `ezbin_seekpos:""`
	Filesize              uint32
	RenderCmdsOffset      uint32
	MaterialsOffset       uint32
	MeshesOffset          uint32
	InvBindMatricesOffset uint32
	_                     [3]byte
	NumBoneMatrices       uint8
	NumMaterials          uint8
	NumMeshes             uint8
	_                     [2]byte
	UpScale               Fixed32
	DownScale             Fixed32
	NumVerts              uint16
	NumPolys              uint16
	NumTris               uint16
	NumQuads              uint16
	BoundingBox           BoundingBox
	_                     [8]uint8
	BoneList              Namelist[uint32]
}

type BoundingBox struct {
	MinX Fixed16
	MinY Fixed16
	MinZ Fixed16
	MaxX Fixed16
	MaxY Fixed16
	MaxZ Fixed16
}

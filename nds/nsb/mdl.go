package nsb

type MDL struct {
	Offset   int64  `ezbin_tell:"namelist_offset"`
	Stamp    string `ezbin_string:"ascii,4"`
	Filesize uint32
	Models   NameList[MDL]
}

type Model struct {
	_        struct{} `ezbin_tell:"model"`
	FileSize uint32

	RenderCmdsOffset      uint32
	Materials             MaterialList `ezbin_offset:"u32"`
	Meshes                MeshList     `ezbin_offset:"u32"`
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
	BoneList    NameList[uint32]
}

type BoundingBox struct {
	MinX float64 `ezbin_fixedpoint:"1,3,12"`
	MinY float64 `ezbin_fixedpoint:"1,3,12"`
	MinZ float64 `ezbin_fixedpoint:"1,3,12"`
	MaxX float64 `ezbin_fixedpoint:"1,3,12"`
	MaxY float64 `ezbin_fixedpoint:"1,3,12"`
	MaxZ float64 `ezbin_fixedpoint:"1,3,12"`
}

type MeshList struct {
	Offset int64 `ezbin_tell:""`
	Meshes NameList[uint32]
}

type MaterialList struct {
	Offset                int64 `ezbin_tell:""`
	TexturePairingsOffset uint32
	PalettePairingsOffset uint32
	Materials             NameList[uint32]
}

package nds

import (
	"fmt"
	"io"

	"github.com/sukus21/nintil/nds/nitrofs"
	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

type header struct {
	GameTitle          string
	GameCode           string
	MakerCode          string
	UnitCode           byte
	EncryptionSeed     byte
	DeviceSize         byte
	RomVersion         byte
	Flags              byte
	Arm9RomOffset      uint32
	Arm9ExecuteAddress uint32
	Arm9Destination    uint32
	Arm9Size           uint32
	Arm7RomOffset      uint32
	Arm7ExecuteAddress uint32
	Arm7Destination    uint32
	Arm7Size           uint32
	FilenameOffset     uint32
	FilenameSize       uint32
	FatOffset          uint32
	FatSize            uint32
	Arm9OverlayOffset  uint32
	Arm9OverlaySize    uint32
	Arm7OverlayOffset  uint32
	Arm7OverlaySize    uint32
	portNormalCommands uint32
	portKeyCommands    uint32
	BannerOffset       uint32
	secureAreaChecksum uint16
	secureAreaDelay    uint16
	Arm9AutoloadList   uint32
	Arm7AutoloadList   uint32
	disableSecureArea  uint64
	RomSize            uint32
	HeaderSize         uint32
	nandRomEnd         uint16
	nandRwStart        uint16
	nintendoLogo       [0x9C]byte
	nintendoLogoCrc    uint16
	HeaderChecksum     uint16
	debugRomOffset     uint32
	debugSize          uint32
	debugRamAddress    uint32
}

func OpenHeader(r io.ReadSeeker) (*header, error) {
	h := &header{}
	strs := make([]byte, 18)
	err := ezbin.Read(r,
		strs,
		&h.UnitCode,
		&h.EncryptionSeed,
		&h.DeviceSize,
		make([]byte, 9),
		&h.RomVersion,
		&h.Flags,

		// Code stuff
		&h.Arm9RomOffset,
		&h.Arm9ExecuteAddress,
		&h.Arm9Destination,
		&h.Arm9Size,
		&h.Arm7RomOffset,
		&h.Arm7ExecuteAddress,
		&h.Arm7Destination,
		&h.Arm7Size,

		// NitroFS data
		&h.FilenameOffset,
		&h.FilenameSize,
		&h.FatOffset,
		&h.FatSize,

		// ARM overlays
		&h.Arm9OverlayOffset,
		&h.Arm9OverlaySize,
		&h.Arm7OverlayOffset,
		&h.Arm7OverlaySize,

		// Misc
		&h.portNormalCommands,
		&h.portKeyCommands,
		&h.BannerOffset,
		&h.secureAreaChecksum,
		&h.secureAreaDelay,
		&h.Arm9AutoloadList,
		&h.Arm7AutoloadList,
		&h.disableSecureArea,
		&h.RomSize,
		&h.HeaderSize,
		make([]byte, 12),
		&h.nandRomEnd,
		&h.nandRwStart,
		make([]byte, 0x28),
		h.nintendoLogo[:],
		&h.nintendoLogoCrc,
		&h.HeaderChecksum,

		// Debug stuff
		&h.debugRomOffset,
		&h.debugSize,
		&h.debugRamAddress,
	)

	// Extract strings
	h.GameTitle = string(strs[:12])
	h.GameCode = string(strs[0x0C:0x10])
	h.MakerCode = string(strs[0x10:0x12])
	return h, err
}

// TODO: validate string lengths
func SaveHeader(w io.WriteSeeker, h *header) error {
	s, _ := w.Seek(0, io.SeekCurrent)
	err := ezbin.Write(w,
		[]byte(h.GameTitle),
		[]byte(h.GameCode),
		[]byte(h.MakerCode),
	)
	if err != nil {
		return err
	}
	if n, _ := w.Seek(0, io.SeekCurrent); n != s+0x12 {
		w.Seek(s, io.SeekStart)
		return fmt.Errorf("save header: title, gamecode or makercode length invalid")
	}

	err = ezbin.Write(w,
		h.UnitCode,
		h.EncryptionSeed,
		h.DeviceSize,
		make([]byte, 9),
		h.RomVersion,
		h.Flags,

		// Code stuff
		h.Arm9RomOffset,
		h.Arm9ExecuteAddress,
		h.Arm9Destination,
		h.Arm9Size,
		h.Arm7RomOffset,
		h.Arm7ExecuteAddress,
		h.Arm7Destination,
		h.Arm7Size,

		// NitroFS data
		h.FilenameOffset,
		h.FilenameSize,
		h.FatOffset,
		h.FatSize,

		// ARM overlays
		h.Arm9OverlayOffset,
		h.Arm9OverlaySize,
		h.Arm7OverlayOffset,
		h.Arm7OverlaySize,

		// Misc
		h.portNormalCommands,
		h.portKeyCommands,
		h.BannerOffset,
		h.secureAreaChecksum,
		h.secureAreaDelay,
		h.Arm9AutoloadList,
		h.Arm7AutoloadList,
		h.disableSecureArea,
		h.RomSize,
		h.HeaderSize,
		make([]byte, 12),
		h.nandRomEnd,
		h.nandRwStart,
		make([]byte, 0x28),
		h.nintendoLogo[:],
		h.nintendoLogoCrc,
		h.HeaderChecksum,

		// Debug stuff (should all be 0)
		h.debugRomOffset,
		h.debugSize,
		h.debugRamAddress,

		// Reserved 0's
		make([]byte, 0x04+0x90+0x0E00),
	)

	// TODO: checksum re-calculations
	return err
}

func (h *header) GetNitroFSInfo() *nitrofs.Info {
	return &nitrofs.Info{
		FntOffset:  h.FilenameOffset,
		FntSize:    h.FilenameSize,
		FatOffset:  h.FatOffset,
		FatSize:    h.FatSize,
		Ovt9Offset: h.Arm9OverlayOffset,
		Ovt9Size:   h.Arm9OverlaySize,
		Ovt7Offset: h.Arm7OverlayOffset,
		Ovt7Size:   h.Arm7OverlaySize,
	}
}

func (h *header) ApplyNitroFSInfo(info *nitrofs.Info) {
	h.FilenameOffset = info.FntOffset
	h.FilenameSize = info.FntSize
	h.FatOffset = info.FatOffset
	h.FatSize = info.FatSize
	h.Arm9OverlayOffset = info.Ovt9Offset
	h.Arm9OverlaySize = info.Ovt9Size
	h.Arm7OverlayOffset = info.Ovt7Offset
	h.Arm7OverlaySize = info.Ovt7Size
}

func (h *header) UpdateChecksum() {
	w := util.NewWriteSeeker(make([]byte, 0x4000))
	SaveHeader(w, h)
	h.HeaderChecksum = CRC16(w.Buf[:0x15E])
}

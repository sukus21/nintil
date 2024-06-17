package nitrofs

import (
	"io"
	"io/fs"

	"github.com/sukus21/nintil/util/ezbin"
)

// I am not sure how overlays work.
// Regardless, now there's an interface to implement if you need them regardless.
type Overlay interface {
	// Returns the address to load the overlay at
	Address() uint32

	// Returns the size in bytes of this overlay
	Size() uint32

	// The file that this overlay points to.
	// The element is never a directory, and never nil.
	//
	// While overlay files belong to the Nitro filesystem,
	// they do not appear in the folder structure at all.
	// TODO: Are there examples of overlays located in the tree?
	Data() []byte

	// Returns start and end of statically initialized data
	StaticData() (start uint32, end uint32)

	// Returns the size of the BSS section for this overlay
	DynamicSize() uint32
}

type OverlaySimple struct {
	// ID of this overlay
	id uint32

	// Address to load overlay at
	loadAddress uint32

	// Size to load
	loadSize uint32

	// size of BSS section for this overlay
	bssSize uint32

	// Staticly initialized data?
	// TODO: what is this?
	stintAddressStart uint32
	stintAddressEnd   uint32

	// Overlay file data
	element fs.File
}

// Reads an overlay directly from the overlay table
func overlayRead(r io.ReaderAt, pos uint32) (uint16, *OverlaySimple, error) {
	o := &OverlaySimple{}
	fileId := uint32(0)
	err := ezbin.ReadAt(r, pos,
		&o.id,
		&o.loadAddress,
		&o.loadSize,
		&o.bssSize,
		&o.stintAddressStart,
		&o.stintAddressEnd,
		&fileId,
		make([]byte, 4), // reserved, filler
	)
	return uint16(fileId), o, err
}

// Write overlay data directly to overlay table
func overlayWrite(w io.Writer, overlay Overlay, overlayId uint32, fileId uint16) error {
	startAddr, endAddr := overlay.StaticData()
	return ezbin.Write(w,
		overlayId,
		overlay.Address(),
		overlay.Size(),
		overlay.DynamicSize(),
		startAddr,
		endAddr,
		uint32(fileId),
		uint32(0), // reserved, filler
	)
}

func (o *OverlaySimple) Address() uint32 {
	return o.loadAddress
}
func (o *OverlaySimple) Size() uint32 {
	return o.loadSize
}
func (o *OverlaySimple) Data() []byte {
	stat, err := o.element.Stat()
	if err != nil {
		return nil
	}
	buf := make([]byte, stat.Size())
	io.ReadFull(o.element, buf)
	return buf
}
func (o *OverlaySimple) StaticData() (uint32, uint32) {
	return o.stintAddressStart, o.stintAddressEnd
}
func (o *OverlaySimple) DynamicSize() uint32 {
	return o.bssSize
}

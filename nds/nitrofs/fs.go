package nitrofs

import (
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
	"github.com/sukus21/nintil/util/mapping"
)

type Info struct {
	FntOffset  uint32
	FntSize    uint32
	FatOffset  uint32
	FatSize    uint32
	Ovt9Offset uint32
	Ovt9Size   uint32
	Ovt7Offset uint32
	Ovt7Size   uint32
}

// This implements fs.FS
type streamFS struct {
	info *Info
	r    util.ReadAtSeeker
	err  error
}

// ----------------------
//
// 	Other public methods
//
// ----------------------

func (blob *streamFS) GetInfo() *Info {
	return blob.info
}
func (blob *streamFS) GetArm9Overlays() []Overlay {
	return blob.readOverlays(blob.info.Ovt9Offset, blob.info.Ovt9Size)
}
func (blob *streamFS) GetArm7Overlays() []Overlay {
	return blob.readOverlays(blob.info.Ovt7Offset, blob.info.Ovt7Size)
}

func (blob *streamFS) readOverlays(offset, size uint32) []Overlay {
	out := make([]Overlay, size/32)

	for i := range out {
		fileId, overlay, err := overlayRead(blob.r, offset+uint32(i)*32)
		if err != nil {
			blob.err = err
			return nil
		}

		element := &streamElement{
			fs:       blob,
			id:       fileId,
			isFolder: false,
		}
		element.open()
		overlay.element = element
		out[i] = overlay
	}

	return out
}

// ---------------------------
//
//  Implement fs.FS
//
// ---------------------------

func (blob *streamFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	root := &streamElement{
		fs:       blob,
		name:     "",
		isFolder: true,
	}

	if name == "." {
		return root, nil
	}

	names := strings.Split(name, "/")
	elem := findByPath(root, names)
	if elem == nil {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrNotExist,
		}
	}

	return elem, nil
}

func findByPath(elem *streamElement, path []string) fs.File {
	if !elem.IsDir() {
		return nil
	}

	// Go through all children
	for _, v := range elem.Children() {
		if v.Name() == path[0] {
			if len(path) == 1 {
				if !v.isFolder {
					v.open()
				}
				return v
			} else {
				return findByPath(v, path[1:])
			}
		}
	}

	// No match found
	return nil
}

// ----------------
//
//  Helper methods
//
// ----------------

func (blob *streamFS) readFilePosition(id uint16) (start uint32, end uint32) {
	if err := ezbin.ReadAt(blob.r, blob.info.FatOffset+uint32(id)*8, &start, &end); err != nil {
		blob.err = err
	}
	return
}

func (blob *streamFS) readContent(id uint16) []byte {
	// Get start and end of file
	fileStart, fileEnd := blob.readFilePosition(id)

	// Read file data
	content := make([]byte, fileEnd-fileStart)
	if _, err := blob.r.ReadAt(content, int64(fileStart)); err != nil {
		blob.err = err
		return nil
	}

	// Return read content
	return content
}

type dirTableEntry struct {
	// offset to subtable, from FNT root
	subtableOffset uint32

	// ID of first file
	firstFile uint16

	// Number of entries OR parent ID
	num uint16
}

// Folder ID must be without the type flag (0x0000..0x0FFF)
func (blob *streamFS) readFolder(id uint16) dirTableEntry {
	var entry dirTableEntry
	ezbin.ReadAt(blob.r, blob.info.FntOffset+uint32(id)*8,
		&entry.subtableOffset,
		&entry.firstFile,
		&entry.num,
	)
	return entry
}

// Get all children for this folder
func (blob *streamFS) getFolderChildren(folderId uint16, from uint32, n int) ([]*streamElement, uint32) {
	folder := blob.readFolder(folderId)
	subtableBase := blob.info.FntOffset + folder.subtableOffset
	blob.r.Seek(int64(subtableBase+from), io.SeekStart)

	all := n <= 0
	var elements []*streamElement
	if all {
		elements = make([]*streamElement, 0, 16)
	} else {
		elements = make([]*streamElement, 0, n)
	}

	fid := folder.firstFile
	for {
		tlen := ezbin.ReadSingle[byte](blob.r)
		if tlen == 0 || tlen == 0x80 {
			at, _ := ezbin.At[uint32](blob.r)
			return elements, at - subtableBase
		}

		isFolder := tlen&0x80 != 0
		tlen &= 0x7F

		// Read name
		fname := make([]byte, tlen)
		if _, err := io.ReadFull(blob.r, fname); err != nil {
			blob.err = err
			at, _ := ezbin.At[uint32](blob.r)
			return elements, at - subtableBase
		}

		if isFolder {
			elements = append(elements, &streamElement{
				fs:       blob,
				name:     string(fname),
				id:       ezbin.ReadSingle[uint16](blob.r) & 0x0FFF,
				isFolder: true,
			})
		} else {
			elements = append(elements, &streamElement{
				fs:       blob,
				name:     string(fname),
				id:       fid,
				isFolder: false,
			})
			fid++
		}

		// In case of limited elements, count down
		if !all {
			n--
			if n == 0 {
				at, _ := ezbin.At[uint32](blob.r)
				return elements, at - subtableBase
			}
		}
	}
}

func (blob *streamFS) updateMapping(mmap *mapping.Mapping) {
	// Do main FS
	fs.WalkDir(blob, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		// Add file to mapping
		streamElement := d.(*streamElement)
		start, end := streamElement.fs.readFilePosition(streamElement.id)
		mmap.AddAt(fmt.Sprintf(mappingFileNamed, streamElement.Name()), start, end-start)
		return nil
	})

	// Do tables
	mmap.AddAt(mappingFAT, blob.info.FatOffset, blob.info.FatSize)
	mmap.AddAt(mappingFNT, blob.info.FntOffset, blob.info.FntSize)
	mmap.AddAt(mappingOVT9, blob.info.Ovt9Offset, blob.info.Ovt9Size)
	mmap.AddAt(mappingOVT7, blob.info.Ovt7Offset, blob.info.Ovt7Size)

	// Do remaining orphan overlay files
	for _, overlay := range append(blob.GetArm9Overlays(), blob.GetArm7Overlays()...) {
		file := overlay.(*OverlaySimple).element.(*streamElement)
		start, end := blob.readFilePosition(file.id)
		mmap.AddAt(fmt.Sprintf(mappingFileUnnamed, file.id), start, end-start)
	}
}

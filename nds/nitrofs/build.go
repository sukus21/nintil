package nitrofs

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"path"
	"slices"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
	"github.com/sukus21/nintil/util/mapping"
)

var ErrNameTooLong = errors.New("file name exceeds 127 byte limit")
var ErrIllegalSymbols = errors.New("file name contains illegal symbols")
var ErrTooLarge = errors.New("filesystem exceeds 512 MB")

func Validate(nfs fs.FS) error {
	_, err := validate(nfs)
	return err
}

type fsCacheFile struct {
	name string
	path string
}

type fsCacheFolder struct {
	name    string
	path    string
	files   []fsCacheFile
	folders []fsCacheFolder
}

type fsCache struct {
	root         fsCacheFolder
	folderLookup map[string]*fsCacheFolder
	numFolders   int
	numFiles     int
	fntSubLen    int
}

func validate(fsys fs.FS) (*fsCache, error) {
	errs := []error{}
	fsc := &fsCache{
		root: fsCacheFolder{
			name: "[root]",
			path: ".",
		},
		folderLookup: map[string]*fsCacheFolder{},
	}
	fsc.folderLookup["."] = &fsc.root
	fsc.numFolders++

	fs.WalkDir(fsys, ".", func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if currentPath == "." {
			return nil
		}

		// Ensure only ascii characters
		illegal := []rune{'\\', '/', '?', '"', '<', '>', '*', ':', ';', '|'}
		for _, v := range d.Name() {
			if v < ' ' || v > '~' || slices.Contains(illegal, v) {
				errs = append(errs, &fs.PathError{
					Op:   "NitroFS-validate",
					Path: currentPath,
					Err:  ErrIllegalSymbols,
				})
				break
			}
		}

		// Ensure name length is within bounds
		encodedName := []byte(d.Name())
		if len(encodedName) > 127 {
			errs = append(errs, &fs.PathError{
				Op:   "NitroFS-validate",
				Path: currentPath,
				Err:  ErrNameTooLong,
			})
		}
		fsc.fntSubLen += len(encodedName) + 1

		// Get parent directory
		parentDir := path.Dir(currentPath)
		parent := fsc.folderLookup[parentDir]

		if d.IsDir() {
			index := len(parent.folders)
			parent.folders = append(parent.folders, fsCacheFolder{
				name: d.Name(),
				path: currentPath,
			})
			fsc.folderLookup[currentPath] = &parent.folders[index]
			fsc.numFolders++
			// Plus 2 for folder ID in parent subtable
			// Plus 1 for null-termination of own subtable
			fsc.fntSubLen += 3
		} else {
			parent.files = append(parent.files, fsCacheFile{
				name: d.Name(),
				path: currentPath,
			})
			fsc.numFiles++
		}

		// Yup, that's pretty much it I think
		return nil
	})

	// Take overlay files into account
	if nfs, ok := fsys.(NitroFS); ok {
		fsc.numFiles += len(nfs.GetArm7Overlays()) + len(nfs.GetArm9Overlays())
	}

	return fsc, errors.Join(errs...)
}

const alignment = uint32(0x0200)

// Builds a NitroFS filesystem from a fs.FS.
// If the given filesystem implements nitrofs.NitroFS, overlay files will be written as well.
func Build(w util.WriteAtSeeker, fsys fs.FS, mmap *mapping.Mapping) (info *Info, err error) {
	defer util.Recover(&err)
	info = &Info{}

	// Is this a valid NitroFS?
	fsc := util.Must1(validate(fsys))

	writeHead := util.Must1(ezbin.At[uint32](w))
	getWriter := func(size uint32, align bool) uint32 {
		if size == 0 {
			return 0
		}
		if align {
			writeHead = ezbin.PadTo(writeHead, alignment)
		}
		pos := writeHead
		writeHead += size
		return pos
	}

	// Initialize FNT (main) writing
	mainSize := uint32(fsc.numFolders) * 8
	info.FntSize = ezbin.PadTo(mainSize+uint32(fsc.fntSubLen), 4)
	info.FntOffset = getWriter(uint32(mainSize), true)
	fntWriter := io.NewOffsetWriter(w, int64(info.FntOffset))

	// Initialize FNT subtable writing
	subOffset := getWriter(info.FntSize-mainSize, false)
	subtableWriter := io.NewOffsetWriter(w, int64(subOffset))

	// Initialize FAT writing
	info.FatSize = uint32(fsc.numFiles) * 8
	info.FatOffset = getWriter(info.FatSize, true)
	fatWriter := io.NewOffsetWriter(w, int64(info.FatOffset))

	// Initialize overlay table writing
	nfs, hasOverlays := fsys.(NitroFS)
	var ovt9Writer *io.OffsetWriter
	var ovt7Writer *io.OffsetWriter
	if hasOverlays {
		info.Ovt9Size = uint32(len(nfs.GetArm9Overlays())) * 32
		info.Ovt9Offset = getWriter(info.Ovt9Size, true)
		ovt9Writer = io.NewOffsetWriter(w, int64(info.Ovt9Offset))

		info.Ovt7Size = uint32(len(nfs.GetArm7Overlays())) * 32
		info.Ovt7Offset = getWriter(info.Ovt7Size, true)
		ovt7Writer = io.NewOffsetWriter(w, int64(info.Ovt7Offset))
	}

	// Initialize file writing
	fileId := uint16(0)
	util.Must1(ezbin.Seek(w, writeHead, io.SeekStart))

	// Handle overlay files
	if hasOverlays {
		for i, ov := range nfs.GetArm9Overlays() {
			util.Must(overlayWrite(ovt9Writer, ov, uint32(i), fileId))
			writeFile(bytes.NewReader(ov.Data()), w, fatWriter)
			fileId++
		}
		for i, ov := range nfs.GetArm7Overlays() {
			util.Must(overlayWrite(ovt7Writer, ov, uint32(i), fileId))
			writeFile(bytes.NewReader(ov.Data()), w, fatWriter)
			fileId++
		}
	}

	type listElem struct {
		folder *fsCacheFolder
		parent uint16
		mine   uint16
	}

	todoFolders := make([]listElem, 0, fsc.numFolders)
	todoFolders = append(todoFolders, listElem{
		folder: &fsc.root,
		parent: uint16(fsc.numFolders),
		mine:   0xF000,
	})
	folderId := uint16(0xF001)

	for len(todoFolders) != 0 {
		listEntry := todoFolders[0]
		folder := listEntry.folder
		parent := listEntry.parent
		myId := listEntry.mine
		todoFolders = todoFolders[1:]

		subtableOffset := util.Must1(ezbin.At[uint32](subtableWriter))
		subtableOffset += mainSize

		// Write main table entry
		util.Must(ezbin.Write(fntWriter, subtableOffset, fileId, parent))

		// Write files
		for i := range folder.files {
			entry := folder.files[i]

			// Open file for reading
			file := util.Must1(fsys.Open(entry.path))
			defer file.Close()
			writeFile(file, w, fatWriter)

			// Write to subtable
			fname := []byte(entry.name)
			util.Must(ezbin.Write(subtableWriter, byte(len(fname)), fname))

			// Increment file ID counter
			fileId++
		}

		// Write folders
		for i := range folder.folders {
			childFolder := &folder.folders[i]

			// Write to subtable
			fname := []byte(childFolder.name)
			util.Must(ezbin.Write(subtableWriter, byte(len(fname)|0x80), fname, folderId))

			// Queue up this folder
			todoFolders = append(todoFolders, listElem{
				folder: childFolder,
				parent: myId,
				mine:   folderId,
			})
			folderId++
		}

		// Terminate sub-table
		util.Must1(subtableWriter.Write([]byte{0}))
	}

	// Ok, looks like that went well
	return
}

// Returns size of written file, and an
func writeFile(dat io.Reader, w io.WriteSeeker, fat io.Writer) {
	// Write file to output
	pos, _ := ezbin.Align(w, alignment)
	n := util.Must1(io.Copy(w, dat))

	// Write entry to FAT
	util.Must(ezbin.Write(fat,
		pos,
		pos+uint32(n),
	))
}

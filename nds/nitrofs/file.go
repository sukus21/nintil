package nitrofs

import (
	"io"
	"io/fs"
	"time"
)

type streamElement struct {
	fs       *streamFS
	r        *io.SectionReader
	name     string
	head     int
	id       uint16
	isFolder bool
}

// -------------------
//
// 	Implement fs.File
//
// -------------------

func (e *streamElement) Stat() (fs.FileInfo, error) {
	return e, nil
}
func (e *streamElement) Read(buf []byte) (int, error) {
	if e.r == nil {
		return 0, fs.ErrClosed
	}

	return e.r.Read(buf)
}
func (e *streamElement) Close() error {
	if e.r == nil {
		return fs.ErrClosed
	}
	e.r = nil
	return nil
}

// --------------------------
//
// 	Implement fs.ReadDirFile
//
// --------------------------

func (e *streamElement) ReadDir(n int) ([]fs.DirEntry, error) {
	if e.head == -1 {
		return nil, fs.ErrClosed
	}

	// Convert children to DirEntry array
	children, offset := e.fs.getFolderChildren(e.id, uint32(e.head), n)
	dirEntries := make([]fs.DirEntry, len(children))
	for i := range children {
		dirEntries[i] = children[i]
	}

	// Not enough elements received
	e.head = int(offset)
	if len(children) != n {
		e.head = -1
		if e.fs.err != nil {
			return dirEntries, e.fs.err
		} else if n > 0 {
			return dirEntries, io.EOF
		}
	}

	// All good
	return dirEntries, nil
}

// -----------------------
//
//  Implement fs.DirEntry
//
// -----------------------

func (e *streamElement) Type() fs.FileMode {
	if e.isFolder {
		return fs.ModeDir
	} else {
		return 0
	}
}
func (e *streamElement) Info() (fs.FileInfo, error) {
	return e, nil
}

// -----------------------
//
// 	Implement fs.FileInfo
//
// -----------------------

func (e *streamElement) Name() string {
	return e.name
}
func (e *streamElement) Size() int64 {
	start, end := e.fs.readFilePosition(e.id)
	return int64(end - start)
}
func (e *streamElement) Mode() fs.FileMode {
	mode := fs.FileMode(0555)
	if e.isFolder {
		mode |= fs.ModeDir
	}
	return mode
}
func (e *streamElement) ModTime() time.Time {
	return time.Time{}
}
func (e *streamElement) IsDir() bool {
	return e.isFolder
}
func (e *streamElement) Sys() any {
	return e.fs
}

// -------------------
//
// 	Other methods
//
// -------------------

// Only valid for folders
func (e *streamElement) Children() []*streamElement {
	if !e.isFolder {
		return nil
	}
	children, _ := e.fs.getFolderChildren(e.id, 0, 0)
	return children
}

// Only valid for non-folders
func (e *streamElement) Data() []byte {
	if e.isFolder {
		return nil
	}
	return e.fs.readContent(e.id)
}

// Only valid for non-folders.
// Initializes reader.
func (e *streamElement) open() {
	start, end := e.fs.readFilePosition(e.id)
	e.r = io.NewSectionReader(e.fs.r, int64(start), int64(end)-int64(start))
}

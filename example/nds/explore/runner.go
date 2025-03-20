package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"reflect"
	"strconv"

	"github.com/sukus21/nintil/compression/lz10"
	"github.com/sukus21/nintil/compression/rlz"
	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

type RunnerFunc func(r *Runner) any

type Runner struct {
	cmdBase   []string
	cmd       []string
	variables map[string]any
	cmdLookup map[string]RunnerFunc
}

// Returns a new runner with the most basic functions
func NewRunner(cmdSrc io.Reader) *Runner {
	r := &Runner{}
	r.variables = make(map[string]any)
	r.cmdLookup = map[string]RunnerFunc{
		// Set variable
		"set": func(r *Runner) any {
			varName := r.CmdRead()
			varVal := ReadArg[any](r)
			r.variables[varName] = varVal
			return nil
		},

		// Get varible
		"get": func(r *Runner) any {
			return r.variables[r.CmdRead()]
		},

		// Decompress LZ10 blob
		"lz10": func(r *Runner) any {
			arg := ReadArg[[]byte](r)
			return util.Must1(lz10.Decompress(arg))
		},

		// Get LZ10 reader
		"lz10r": func(r *Runner) any {
			arg := ReadArg[io.Reader](r)
			return util.Must1(lz10.NewReader(arg))
		},

		// Decompress RLZ blob
		"rlz": func(r *Runner) any {
			arg := ReadArg[[]byte](r)
			return util.Must1(rlz.Decompress(arg))
		},

		// Get RLZ reader
		"rlzr": func(r *Runner) any {
			arg := ReadArg[io.Reader](r)
			return util.Must1(rlz.NewReader(arg))
		},

		// Convert blob to reader
		"r": func(r *Runner) any {
			arg := ReadArg[[]byte](r)
			return bytes.NewReader(arg)
		},

		// Convert reader to blob
		"b": func(r *Runner) any {
			arg := ReadArg[io.Reader](r)
			return util.Must1(io.ReadAll(arg))
		},

		// Prints whatever
		"print": func(r *Runner) any {
			fmt.Println(ReadArg[any](r))
			return nil
		},

		// Get length of thing
		"len": func(r *Runner) any {
			value := ReadArg[any](r)
			refValue := reflect.ValueOf(value)
			refValue.Kind()
			return refValue.Len()
		},

		// Export image as PNG
		// - string				fname
		// - image.Image		img
		"png": func(r *Runner) any {
			fname := ReadArg[string](r)
			img := ReadArg[image.Image](r)

			f := util.Must1(os.Create(fname))
			defer f.Close()
			util.Must(png.Encode(f, img))

			return nil
		},

		// Read contents of file
		// - string				fname
		// + []byte
		"fread": func(r *Runner) any {
			fname := ReadArg[string](r)
			return util.Must1(os.ReadFile(fname))
		},

		// Write contents to file
		// - string				fname
		// - []byte				content
		"fwrite": func(r *Runner) any {
			fname := ReadArg[string](r)
			content := ReadArg[[]byte](r)
			util.Must(os.WriteFile(fname, content, os.ModePerm))
			return nil
		},

		// Set current working directory
		// - string				dname
		"cwd": func(r *Runner) any {
			cwd := ReadArg[string](r)
			os.Chdir(cwd)
			return nil
		},
	}

	return r
}

func (r *Runner) CmdAppend(cmd []string) {
	diff := len(r.cmdBase) - len(r.cmd)
	r.cmdBase = append(r.cmdBase, cmd...)
	r.cmd = r.cmdBase[diff:]
}

func (r *Runner) CmdPrepend(cmd []string) {
	diff := len(r.cmdBase) - len(r.cmd)
	diff2 := diff - len(cmd)
	if diff2 >= 0 && diff >= len(cmd) {
		copy(r.cmdBase[diff2:], cmd)
		return
	}

	r.cmdBase = append(cmd, r.cmd...)
	r.cmd = r.cmdBase
}

func (r *Runner) CmdRun() (err error) {
	defer util.Recover(&err)

	for len(r.cmd) > 0 {
		cmdEntry := r.CmdRead()
		cmdFunc, ok := r.cmdLookup[cmdEntry]
		if !ok {
			return errors.New("unknown command " + cmdEntry)
		}

		cmdFunc(r)
	}

	return nil
}

func (r *Runner) CmdRead() string {
	val := r.cmd[0]
	r.cmd = r.cmd[1:]
	return val
}

func ReadArg[T any](r *Runner) T {
	cmd := r.CmdRead()

	// Is this a function
	if cmdFunc, ok := r.cmdLookup[cmd]; ok {
		return resolveNumber[T](cmdFunc(r))
	}
	if cmdVar, ok := r.variables[cmd]; ok {
		return resolveNumber[T](cmdVar)
	}

	// Is it a number?
	if n, err := strconv.ParseInt(cmd, 0, 64); err == nil {
		return resolveNumber[T](n)
	}

	// Ok, I give up. Return stringe
	return any(cmd).(T)
}

func resolveNumber[T any](n any) T {
	if val, ok := any(n).(T); ok {
		return val
	}

	num := getNumber(n)

	var out T
	switch e := any(out).(type) {
	case int64:
		return any(n).(T)
	case uint32:
		return any(uint32(num)).(T)
	case byte:
		return any(uint16(num)).(T)
	case int:
		return any(int(num)).(T)
	case bool:
		return any(num != 0).(T)
	case any:
		_ = e
		return any(n).(T)
	default:
		panic("this number type is no good :(")
	}
}

func getNumber[T any](n T) int64 {
	switch n := any(n).(type) {
	case int:
		return int64(n)
	case uint:
		return int64(n)
	case uintptr:
		return int64(n)
	case uint8:
		return int64(n)
	case int8:
		return int64(n)
	case uint16:
		return int64(n)
	case int16:
		return int64(n)
	case uint32:
		return int64(n)
	case int32:
		return int64(n)
	case uint64:
		return int64(n)
	case int64:
		return int64(n)
	default:
		panic("bad number type :(")
	}
}

func ReadVar[T any](r *Runner, name string) T {
	varVal, ok := r.variables[name]
	if !ok {
		panic("variable not set before reading")
	}
	return resolveNumber[T](varVal)
}

func ReadVarIf[T any](r *Runner, name string) T {
	var empty T
	if _, ok := r.variables[name]; !ok {
		return empty
	}
	return ReadVar[T](r, name)
}

func (t *Runner) AddFunc(name string, f RunnerFunc) {
	t.cmdLookup[name] = f
}

func drawImgBitmap(w int, h int, t4bpp bool, gfx *bytes.Reader, pal color.Palette) image.Image {
	img := image.NewPaletted(
		image.Rect(0, 0, w, h),
		pal,
	)

	tsize := int(gfx.Size())
	if t4bpp {
		tsize *= 2
	}
	x := 0
	y := 0
	lastPix := byte(0)
	for i := 0; i < tsize && y < h; i++ {
		pix := lastPix >> 4
		if i&1 != 0 || !t4bpp {
			pix, _ = gfx.ReadByte()
			lastPix = pix
			if t4bpp {
				pix &= 15
			}
		}
		img.SetColorIndex(x, y, pix)

		x++
		if x >= w {
			x = 0
			y++
		}
	}

	return img
}

func drawImgTilemap(w int, h int, tls []nds.Tile, tlm *bytes.Reader, pal color.Palette) image.Image {
	img := image.NewPaletted(
		image.Rect(0, 0, w*8, h*8),
		pal,
	)

	xt := 0
	yt := 0
	for i := 0; i < int(tlm.Size())/2 && yt < h; i++ {
		raw := ezbin.ReadSingle[uint16](tlm)
		tileId := raw & 0x3FF
		tileMirror := raw&0x0400 != 0
		tileFlip := raw&0x0800 != 0
		// tilePal := int(raw >> 12)

		tile := &tls[tileId]
		x := xt * 8
		y := yt * 8
		xt++
		if xt >= w {
			xt = 0
			yt++
		}
		nds.DrawTileSamePalette(
			img, tile,
			x, y,
			tileMirror, tileFlip,
		)
	}

	return img
}

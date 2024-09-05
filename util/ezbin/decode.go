package ezbin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strconv"

	"github.com/sukus21/nintil/util"
)

func Decode[T any](src io.Reader) (v T, err error) {
	defer util.Recover(&err)

	r := &EndianedReader{
		Reader:    src,
		ByteOrder: getEndian(src),
	}

	// Simple things
	t := reflect.New(reflect.TypeFor[T]()).Elem()
	decodeValue(t, r, defaultTags)
	v = t.Interface().(T)
	return
}

const tagLength = "ezbin_length"
const tagString = "ezbin_string"
const tagByteorder = "ezbin_byteorder"
const tagSeekpos = "ezbin_seekpos"

type decodeTags struct {
	length  int
	xstring string
}

var defaultTags = decodeTags{
	length: -1,
}

var ErrSliceMissingLength = errors.New("cannot decode slice without length tag")
var ErrUnknownLengthProperty = errors.New("field specified in length tag is not yet set or does not exist")
var ErrInvalidByteOrderType = errors.New("invalid byte order type, expected [2]byte")
var ErrInvalidByteOrder = errors.New("invalid byte order, expected 0xFEFF or 0xFFFE")
var ErrUnknownString = errors.New("unknown string type")
var ErrNotSeeker = errors.New("reader does not implement io.Seeker")
var ErrInvalidSeekposType = errors.New("invalid seekpos type, expected integer")

var SignatureLE = [2]byte{0xFF, 0xFE}
var SignatureBE = [2]byte{0xFE, 0xFF}

func decodeValue(t reflect.Value, r *EndianedReader, tags decodeTags) {
	switch t.Kind() {

	// Basic kinds
	case reflect.Uint8:
		t.SetUint(uint64(ReadSingle[uint8](r)))
	case reflect.Uint16:
		t.SetUint(uint64(ReadSingle[uint16](r)))
	case reflect.Uint32:
		t.SetUint(uint64(ReadSingle[uint32](r)))
	case reflect.Uint64:
		t.SetUint(uint64(ReadSingle[uint64](r)))
	case reflect.Int8:
		t.SetInt(int64(ReadSingle[int8](r)))
	case reflect.Int16:
		t.SetInt(int64(ReadSingle[int16](r)))
	case reflect.Int32:
		t.SetInt(int64(ReadSingle[int32](r)))
	case reflect.Int64:
		t.SetInt(int64(ReadSingle[int64](r)))
	case reflect.Float32:
		t.SetFloat(float64(ReadSingle[float32](r)))
	case reflect.Float64:
		t.SetFloat(float64(ReadSingle[float64](r)))

	// Special case: read byte order
	case reflect.Interface:
		typ := t.Type()
		switch {
		case typ.Implements(reflect.TypeFor[binary.ByteOrder]()):
			t2 := reflect.New(reflect.TypeFor[[2]byte]()).Elem()
			decodeValue(t2, r, tags)
			newEndian := GetEndianFromSignature(t2.Interface().([2]byte))
			t.Set(reflect.ValueOf(newEndian))

		default:
			panic("cannot deserialize this kind of type")
		}

	// Arrays and slices
	case reflect.Array:
		decodeArray(t, r)
	case reflect.Slice:
		if tags.length < 0 {
			panic(ErrSliceMissingLength)
		}
		t.Set(reflect.MakeSlice(t.Type(), tags.length, tags.length))
		decodeArray(t, r)

	// Other weird ones
	case reflect.Bool:
		t.SetBool(ReadSingle[byte](r) != 0)
	case reflect.Pointer:
		val := reflect.New(t.Type())
		decodeValue(val.Elem(), r, tags)
		t.Set(val)
	case reflect.String:
		decodeString(t, r, tags)

	// Structs
	case reflect.Struct:
		decodeStruct(t, r)

	default:
		panic("cannot deserialize this kind of type")
	}
}

func decodeArray(t reflect.Value, r *EndianedReader) {
	elems := t.Len()
	for i := 0; i < elems; i++ {
		decodeValue(t.Index(i), r, defaultTags)
	}
}

func decodeStruct(t reflect.Value, r *EndianedReader) {
	numFields := t.NumField()
	seenInts := make(map[string]int)

	for i := 0; i < numFields; i++ {
		fieldType := t.Type().Field(i)
		field := t.Field(i)

		// Seek tag?
		if _, ok := fieldType.Tag.Lookup(tagSeekpos); ok {
			seeker, ok := r.Reader.(io.Seeker)
			if !ok {
				panic(ErrNotSeeker)
			}
			at := util.Must1(At[int64](seeker))
			if field.CanUint() {
				field.SetUint(uint64(at))
			} else if field.CanInt() {
				field.SetInt(at)
			} else {
				panic(ErrInvalidSeekposType)
			}
			continue
		}

		// Read tags
		tags := defaultTags
		if lengthField, ok := fieldType.Tag.Lookup(tagLength); ok {
			if intval, err := strconv.ParseInt(lengthField, 0, 64); err == nil {
				tags.length = int(intval)
			} else if length, ok := seenInts[lengthField]; ok {
				tags.length = length
			} else {
				panic(ErrUnknownLengthProperty)
			}
		}
		if stringType, ok := fieldType.Tag.Lookup(tagString); ok {
			tags.xstring = stringType
		}

		// Decode value of field
		decodeValue(field, r, tags)

		// Save integer, maybe
		indir := reflect.Indirect(field)
		if indir.CanInt() {
			seenInts[fieldType.Name] = int(indir.Int())
		} else if indir.CanUint() {
			seenInts[fieldType.Name] = int(indir.Uint())
		}

		// Set byte order?
		if _, ok := fieldType.Tag.Lookup(tagByteorder); ok {
			if bytes, ok := indir.Interface().([2]byte); ok {
				r.ByteOrder = GetEndianFromSignature(bytes)
			} else if bo, ok := indir.Interface().(binary.ByteOrder); ok {
				r.ByteOrder = bo
			} else {
				panic(ErrInvalidByteOrderType)
			}
		}
	}
}

func decodeString(t reflect.Value, r *EndianedReader, tags decodeTags) {
	str := ""

	switch tags.xstring {
	case "ascii":
		var sr io.Reader = r
		if tags.length >= 0 {
			dat := make([]byte, tags.length)
			util.Must1(io.ReadFull(r, dat))
			sr = bytes.NewReader(dat)
		}
		for {
			b := ReadSingle[byte](sr)
			if b != 0 {
				str += string(rune(b))
			} else {
				break
			}
		}

	// TODO: support more string types
	default:
		panic(ErrUnknownString)
	}

	t.SetString(str)
}

func GetEndianFromSignature(signature [2]byte) binary.ByteOrder {
	if signature == SignatureBE {
		return binary.BigEndian
	} else if signature == SignatureLE {
		return binary.LittleEndian
	} else {
		panic(ErrInvalidByteOrder)
	}
}

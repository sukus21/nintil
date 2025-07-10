package ezbin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"

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
const tagFixedPoint = "ezbin_fixedpoint"

type decodeTags struct {
	length  int
	xstring string
	fixed   string
}

var defaultTags = decodeTags{
	length: -1,
}

var ErrSliceMissingLength = errors.New("cannot decode slice without length tag")
var ErrUnknownLengthProperty = errors.New("field specified in length tag is not yet set or does not exist")
var ErrInvalidByteOrderType = errors.New("invalid byte order type, expected [2]byte")
var ErrInvalidByteOrder = errors.New("invalid byte order, expected 0xFEFF or 0xFFFE")
var ErrUnknownString = errors.New("unknown string type")
var ErrInvalidFixedPointSignature = errors.New("invalid fixed point tag, expected \"int,int,int\"")
var ErrInvalidFixedPointSize = errors.New("invalid fixed point size, expected multiple of 8 bits, max 64")
var ErrInvalidFixedPointBits = errors.New("invalid fixed point bits, bit counts cannot be below 0")
var ErrInvalidFixedPointSign = errors.New("invalid fixed point sign, sign can only be 0 or 1 bit")
var ErrInvalidFixedPointType = errors.New("invalid fixed point type, expected float32 or float64")
var ErrNotSeeker = errors.New("reader does not implement io.Seeker")
var ErrInvalidSeekposType = errors.New("invalid seekpos type, expected integer")

var SignatureLE = [2]byte{0xFF, 0xFE}
var SignatureBE = [2]byte{0xFE, 0xFF}

func decodeValue(t reflect.Value, r *EndianedReader, tags decodeTags) {
	switch t.Kind() {

	// Basic kinds
	case reflect.Uint8:
		val := uint64(ReadSingle[uint8](r))
		if t.CanSet() {
			t.SetUint(val)
		}
	case reflect.Uint16:
		val := uint64(ReadSingle[uint16](r))
		if t.CanSet() {
			t.SetUint(val)
		}
	case reflect.Uint32:
		val := uint64(ReadSingle[uint32](r))
		if t.CanSet() {
			t.SetUint(val)
		}
	case reflect.Uint64:
		val := uint64(ReadSingle[uint64](r))
		if t.CanSet() {
			t.SetUint(val)
		}
	case reflect.Int8:
		val := int64(ReadSingle[int8](r))
		if t.CanSet() {
			t.SetInt(val)
		}
	case reflect.Int16:
		val := int64(ReadSingle[int16](r))
		if t.CanSet() {
			t.SetInt(val)
		}
	case reflect.Int32:
		val := int64(ReadSingle[int32](r))
		if t.CanSet() {
			t.SetInt(val)
		}
	case reflect.Int64:
		val := int64(ReadSingle[int64](r))
		if t.CanSet() {
			t.SetInt(val)
		}
	case reflect.Float32:
		val := decodeFloat[float32](r, tags)
		if t.CanSet() {
			t.SetFloat(val)
		}
	case reflect.Float64:
		val := decodeFloat[float64](r, tags)
		if t.CanSet() {
			t.SetFloat(val)
		}

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
			panic("cannot decode this kind of type")
		}

	// Arrays and slices
	case reflect.Array:
		decodeArray(t, r, tags)
	case reflect.Slice:
		if tags.length < 0 {
			panic(ErrSliceMissingLength)
		}
		val := reflect.MakeSlice(t.Type(), tags.length, tags.length)

		tags.length = defaultTags.length
		decodeArray(val, r, tags)
		if t.CanSet() {
			t.Set(val)
		}

	// Other weird ones
	case reflect.Bool:
		val := ReadSingle[byte](r) != 0
		if t.CanSet() {
			t.SetBool(val)
		}
	case reflect.Pointer:
		val := reflect.New(t.Type())
		decodeValue(val.Elem(), r, tags)
		if t.CanSet() {
			t.Set(val)
		}
	case reflect.String:
		decodeString(t, r, tags)

	// Structs
	case reflect.Struct:
		decodeStruct(t, r)

	default:
		panic("cannot deserialize this kind of type")
	}
}

func decodeArray(t reflect.Value, r *EndianedReader, tags decodeTags) {
	elems := t.Len()
	for i := range elems {
		decodeValue(t.Index(i), r, tags)
	}
}

func decodeStruct(t reflect.Value, r *EndianedReader) {
	numFields := t.NumField()
	seenInts := make(map[string]int)

	for i := range numFields {
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
		if fixedType, ok := fieldType.Tag.Lookup(tagFixedPoint); ok {
			tags.fixed = fixedType
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

	if t.CanSet() {
		t.SetString(str)
	}
}

func decodeFloat[T float32 | float64](r *EndianedReader, tags decodeTags) float64 {
	if tags.fixed != "" {
		return decodeFixedPoint(r, tags.fixed)
	} else {
		return float64(ReadSingle[T](r))
	}
}

func decodeFixedPoint(r *EndianedReader, fixedTag string) float64 {
	bits := strings.Split(fixedTag, ",")
	if len(bits) != 3 {
		panic(ErrInvalidFixedPointSignature)
	}

	// Get bit counts
	totalBits := 0
	bitCounts := [3]int{}
	for i := range bitCounts {
		bitCount, err := strconv.ParseInt(bits[i], 0, 64)
		if err != nil {
			panic(ErrInvalidFixedPointSignature)
		}

		bitCounts[i] = int(bitCount)
		totalBits += int(bitCount)
	}

	// Validate bit count in fixed-point field
	if totalBits&0x7 != 0 || totalBits > 64 {
		panic(ErrInvalidFixedPointSize)
	}
	if bitCounts[0] != 0 && bitCounts[0] != 1 {
		panic(ErrInvalidFixedPointSign)
	}
	if bitCounts[1] < 0 || bitCounts[2] < 0 {
		panic(ErrInvalidFixedPointBits)
	}

	// Read bits
	rawData := uint64(0)
	rawBytes := make([]byte, totalBits/8)
	util.Must1(r.ReadWithOrder(rawBytes))
	for i := range rawBytes {
		rawData |= uint64(rawBytes[i]) << ((len(rawBytes) - (i + 1)) * 8)
	}

	// Get number parts
	negative := bitCounts[0] == 1 && (rawData&(1<<(totalBits-1))) != 0
	if negative {
		rawData ^= (1 << totalBits) - 1
		rawData += 1
		bitCounts[1] += 1
	}

	fraction := rawData & ((1 << bitCounts[2]) - 1)
	whole := (rawData >> bitCounts[2]) & ((1 << bitCounts[1]) - 1)

	// Convert to float
	final := float64(whole)
	final += float64(fraction) / float64(int(1)<<bitCounts[2])
	if negative {
		final = -final
	}

	return final
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

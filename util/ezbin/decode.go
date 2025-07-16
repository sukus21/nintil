package ezbin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"maps"
	"reflect"
	"strconv"
	"strings"

	"github.com/sukus21/nintil/util"
)

type Decodable interface {
	// Reader may be an io.Seeker as well
	EzbinDecode(r io.Reader) error
}

const tagLength = "ezbin_length"
const tagString = "ezbin_string"
const tagByteorder = "ezbin_byteorder"
const tagFixedPoint = "ezbin_fixedpoint"
const tagTell = "ezbin_tell"
const tagOffset = "ezbin_offset"
const tagOffsetArray = "ezbin_offset_array"

var ErrSliceMissingLength = errors.New("cannot decode slice without length tag")
var ErrUnknownNamedInt = errors.New("malformed int or invalid named int field")
var ErrInvalidByteOrderType = errors.New("invalid byte order type, expected [2]byte")
var ErrInvalidByteOrder = errors.New("invalid byte order, expected 0xFEFF or 0xFFFE")
var ErrUnknownString = errors.New("unknown string type")
var ErrInvalidFixedPointSignature = errors.New("invalid fixed point tag, expected \"int,int,int\"")
var ErrInvalidFixedPointSize = errors.New("invalid fixed point size, expected multiple of 8 bits, max 64")
var ErrInvalidFixedPointBits = errors.New("invalid fixed point bits, bit counts cannot be below 0")
var ErrInvalidFixedPointSign = errors.New("invalid fixed point sign, sign can only be 0 or 1 bit")
var ErrInvalidFixedPointType = errors.New("invalid fixed point type, expected float32 or float64")
var ErrNotSeeker = errors.New("reader does not implement io.Seeker")
var ErrInvalidTellType = errors.New("invalid seekpos type, expected integer")

var SignatureLE = [2]byte{0xFF, 0xFE}
var SignatureBE = [2]byte{0xFE, 0xFF}

func Decode[T any](r io.Reader) (v T, err error) {
	defer util.Recover(&err)

	d := newDecoder(r)

	// Simple things
	t := reflect.New(reflect.TypeFor[T]()).Elem()
	d.decodeValue(t, "")
	v = t.Interface().(T)
	return
}

type decoder struct {
	*EndianedReader                  // Source file
	io.Seeker                        // The seeker of reader's internal reader, may be nil
	namedInts       map[string]int64 // Named ints
	offset          int64            // Struct start offset
}

func newDecoder(r io.Reader) (out decoder) {
	out.EndianedReader = &EndianedReader{
		Reader:    r,
		ByteOrder: getEndian(r),
	}

	out.namedInts = make(map[string]int64)
	out.Seeker, _ = r.(io.Seeker)
	if out.Seeker != nil {
		out.offset, _ = At[int64](out.Seeker)
		out.namedInts["root"] = out.offset
	}

	return
}

func (d *decoder) clone() *decoder {
	c := new(decoder)
	*c = *d

	c.namedInts = maps.Clone(d.namedInts)
	if c.Seeker != nil {
		c.offset, _ = At[int64](d)
	}

	return c
}

var typeofDecodable = reflect.TypeFor[Decodable]()

// Tries to read an int.
// It tries the following, in this order:
//
// 1. Parse name as int
// 2. Get int from named fields
// 3. Read int from stream using named type
//
// If all of these fail, the method panics
func (d *decoder) getInt(intName string) int64 {
	if intval, err := strconv.ParseInt(intName, 0, 64); err == nil {
		return intval
	} else if seenInt, ok := d.namedInts[intName]; ok {
		return seenInt
	} else if intVal, ok := d.readIntFromNamedType(intName); ok {
		return intVal
	} else {
		panic(ErrUnknownNamedInt)
	}
}

func (d *decoder) decodeValue(t reflect.Value, tags reflect.StructTag) {
	// Offset tag, move position
	if offsetType, ok := tags.Lookup(tagOffset); ok {
		tags = StructTagRemove(tags, tagOffset)
		if d.Seeker == nil {
			panic(ErrNotSeeker)
		}
		offsetParts := strings.SplitN(offsetType, ",", 2)

		// Get data offset
		offsetBase := d.offset
		offset := d.getInt(offsetParts[0])
		if len(offsetParts) == 2 {
			offsetBase = d.getInt(offsetParts[1])
		}

		// Save current offset and set new offset
		posBefore := util.Must1(At[int64](d.Seeker))

		// Seek, read, reset position, and return
		util.Must1(d.Seeker.Seek(offsetBase+offset, io.SeekStart))
		d.decodeValue(t, tags)
		util.Must1(d.Seeker.Seek(posBefore, io.SeekStart))
		return
	}

	// Does subject implement the thing?
	if t.Type().Implements(typeofDecodable) {
		nv := reflect.New(t.Type()).Elem()
		iface := nv.Interface().(Decodable)
		util.Must(iface.EzbinDecode(d))
		t.Set(nv)
		return
	}

	if reflect.PointerTo(t.Type()).Implements(typeofDecodable) {
		nvp := reflect.New(t.Type())
		iface := nvp.Interface().(Decodable)
		util.Must(iface.EzbinDecode(d))
		t.Set(nvp.Elem())
		return
	}

	switch t.Kind() {

	// Basic kinds
	case reflect.Uint8:
		val := uint64(ReadSingle[uint8](d))
		if t.CanSet() {
			t.SetUint(val)
		}
	case reflect.Uint16:
		val := uint64(ReadSingle[uint16](d))
		if t.CanSet() {
			t.SetUint(val)
		}
	case reflect.Uint32:
		val := uint64(ReadSingle[uint32](d))
		if t.CanSet() {
			t.SetUint(val)
		}
	case reflect.Uint64:
		val := uint64(ReadSingle[uint64](d))
		if t.CanSet() {
			t.SetUint(val)
		}
	case reflect.Int8:
		val := int64(ReadSingle[int8](d))
		if t.CanSet() {
			t.SetInt(val)
		}
	case reflect.Int16:
		val := int64(ReadSingle[int16](d))
		if t.CanSet() {
			t.SetInt(val)
		}
	case reflect.Int32:
		val := int64(ReadSingle[int32](d))
		if t.CanSet() {
			t.SetInt(val)
		}
	case reflect.Int64:
		val := int64(ReadSingle[int64](d))
		if t.CanSet() {
			t.SetInt(val)
		}
	case reflect.Float32:
		val := decodeFloat[float32](d, tags)
		if t.CanSet() {
			t.SetFloat(val)
		}
	case reflect.Float64:
		val := decodeFloat[float64](d, tags)
		if t.CanSet() {
			t.SetFloat(val)
		}

	// Special case: read byte order
	case reflect.Interface:
		typ := t.Type()
		switch {
		case typ.Implements(reflect.TypeFor[binary.ByteOrder]()):
			t2 := reflect.New(reflect.TypeFor[[2]byte]()).Elem()
			d.decodeValue(t2, tags)
			newEndian := GetEndianFromSignature(t2.Interface().([2]byte))
			t.Set(reflect.ValueOf(newEndian))

			// Set byte order?
			if _, ok := tags.Lookup(tagByteorder); ok {
				d.ByteOrder = newEndian
			}

		default:
			panic("cannot decode this kind of type")
		}

	// Arrays and slices
	case reflect.Array:
		d.decodeArray(t, tags)
	case reflect.Slice:
		d.decodeSlice(t, tags)

	// Other weird ones
	case reflect.Bool:
		val := ReadSingle[byte](d) != 0
		if t.CanSet() {
			t.SetBool(val)
		}
	case reflect.Pointer:
		val := reflect.New(t.Type().Elem())
		d.decodeValue(val.Elem(), tags)
		if t.CanSet() {
			t.Set(val)
		}
	case reflect.String:
		d.decodeString(t, tags)

	// Structs
	case reflect.Struct:
		d.decodeStruct(t)

	default:
		panic("cannot deserialize this kind of type")
	}
}

func (d *decoder) decodeSlice(t reflect.Value, tags reflect.StructTag) {
	lengthTag, ok := tags.Lookup(tagLength)
	if !ok {
		panic(ErrSliceMissingLength)
	}

	// Parse length tag
	lengthParts := strings.SplitN(lengthTag, ",", 2)
	length := int(d.getInt(lengthParts[0]))

	// Include another part?
	if len(lengthParts) == 2 {
		// Sub-type MUST be byte for this to work
		contentKind := t.Type().Elem().Kind()
		if contentKind != reflect.Uint8 && contentKind != reflect.Int8 {
			panic("length-origin can only be applied to bytes")
		}

		lengthBase := int(d.getInt(lengthParts[1]))
		here, _ := At[int](d)
		length -= here - lengthBase
	}

	// Read data
	val := reflect.MakeSlice(t.Type(), length, length)
	d.decodeArray(val, StructTagRemove(tags, tagLength))
	if t.CanSet() {
		t.Set(val)
	}
}

func (d *decoder) decodeArray(t reflect.Value, tags reflect.StructTag) {
	// Array offset tag?
	memberTags := tags
	if arrayOffset, ok := tags.Lookup(tagOffsetArray); ok {
		memberTags = StructTagRemove(memberTags, tagOffsetArray)
		memberTags = StructTagRemove(memberTags, tagOffset)
		memberTags = StructTagAdd(memberTags, tagOffset, arrayOffset)
	}

	// Decode array
	elems := t.Len()
	for i := range elems {
		d.decodeValue(t.Index(i), memberTags)
	}

	// Set byte order?
	if _, ok := tags.Lookup(tagByteorder); ok {
		indir := reflect.Indirect(t)
		if bytes, ok := indir.Interface().([2]byte); ok {
			d.ByteOrder = GetEndianFromSignature(bytes)
		} else {
			panic(ErrInvalidByteOrderType)
		}
	}
}

func (d *decoder) readIntFromNamedType(intTypeName string) (int64, bool) {
	// Read int from file instead
	intKind, ok := intKindFromName(intTypeName)
	if !ok {
		return 0, false
	}
	intValue := reflect.New(intTypeFromKind(intKind)).Elem()
	d.decodeValue(intValue, "")

	if intValue.CanInt() {
		return int64(intValue.Int()), true
	} else if intValue.CanUint() {
		return int64(intValue.Uint()), true
	} else {
		panic("cannot get int, because int type cannot int or uint. huh???")
	}
}

func (d *decoder) decodeStruct(t reflect.Value) {
	d = d.clone()

	numFields := t.NumField()
	for i := range numFields {
		fieldType := t.Type().Field(i)
		field := t.Field(i)

		d.decodeField(field, fieldType)
	}
}

func (d *decoder) decodeField(field reflect.Value, fieldType reflect.StructField) {
	// Tell tag?
	if tellName, ok := fieldType.Tag.Lookup(tagTell); ok {
		d.decodeTell(field, tellName)
		return
	}

	// Decode value of field
	d.decodeValue(field, fieldType.Tag)

	// Save integer, maybe
	indir := reflect.Indirect(field)
	if indir.CanInt() {
		d.namedInts[fieldType.Name] = indir.Int()
	} else if indir.CanUint() {
		d.namedInts[fieldType.Name] = int64(indir.Uint())
	}
}

func (d *decoder) decodeTell(field reflect.Value, tellName string) {
	if d.Seeker == nil {
		panic(ErrNotSeeker)
	}
	offset := util.Must1(At[int64](d.Seeker))

	// Set field, if possible
	if field.CanUint() {
		field.SetUint(uint64(offset))
		d.namedInts[field.Type().Name()] = offset
	} else if field.CanInt() {
		field.SetInt(offset)
		d.namedInts[field.Type().Name()] = offset
	}

	// Register specified tell name
	if tellName != "" {
		d.namedInts[tellName] = offset
	}
}

func (d *decoder) decodeString(t reflect.Value, tags reflect.StructTag) {

	// Validate string tag
	stringTag, ok := tags.Lookup(tagString)
	if !ok {
		panic("no string tag provided")
	}
	parts := strings.SplitN(stringTag, ",", 2)
	if len(parts) == 0 {
		panic("malformed string tag")
	}

	stringType := parts[0]

	// Do we read string from a fixed-length buffer?
	stringReader := io.Reader(d)
	if len(parts) == 2 {
		stringLength := d.getInt(parts[1])
		if stringLength < 0 {
			panic("string cannot have length < 0")
		}

		dat := make([]byte, stringLength)
		util.Must1(io.ReadFull(d, dat))
		stringReader = bytes.NewReader(dat)
	}

	str := ""
	switch stringType {
	case "ascii":
		for {
			b := ReadSingle[byte](stringReader)
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

// Go I'm begging you, PLEASE let me make generic methods!
func decodeFloat[T float32 | float64](d *decoder, tags reflect.StructTag) float64 {
	if fixedType, ok := tags.Lookup(tagFixedPoint); ok {
		return d.decodeFixedPoint(fixedType)
	} else {
		return float64(ReadSingle[T](d))
	}
}

func (d *decoder) decodeFixedPoint(fixedTag string) float64 {
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
	util.Must1(d.ReadWithOrder(rawBytes))
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
	switch signature {
	case SignatureBE:
		return binary.BigEndian
	case SignatureLE:
		return binary.LittleEndian
	default:
		panic(ErrInvalidByteOrder)
	}
}

func intKindFromName(name string) (reflect.Kind, bool) {
	switch strings.ToLower(name) {
	case "u8", "uint8":
		return reflect.Uint8, true
	case "u16", "uint16":
		return reflect.Uint16, true
	case "u32", "uint32":
		return reflect.Uint32, true
	case "u64", "uint64":
		return reflect.Uint64, true

	case "s8", "i8", "int8":
		return reflect.Int8, true
	case "s16", "i16", "int16":
		return reflect.Int16, true
	case "s32", "i32", "int32":
		return reflect.Int32, true
	case "s64", "i64", "int64":
		return reflect.Int64, true

	default:
		return reflect.Invalid, false
	}
}

func intTypeFromKind(kind reflect.Kind) reflect.Type {
	switch kind {
	case reflect.Uint8:
		return reflect.TypeFor[uint8]()
	case reflect.Uint16:
		return reflect.TypeFor[uint16]()
	case reflect.Uint32:
		return reflect.TypeFor[uint32]()
	case reflect.Uint64:
		return reflect.TypeFor[uint64]()
	case reflect.Int8:
		return reflect.TypeFor[int8]()
	case reflect.Int16:
		return reflect.TypeFor[int16]()
	case reflect.Int32:
		return reflect.TypeFor[int32]()
	case reflect.Int64:
		return reflect.TypeFor[int64]()

	default:
		panic("non-binary int type")
	}
}

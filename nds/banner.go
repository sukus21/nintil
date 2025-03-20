package nds

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"
	"unicode/utf16"

	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
)

const (
	BannerVersionOriginal = 0x0001
	BannerVersionChinese  = 0x0002
	BannerVersionKorean   = 0x0003
	BannerVersionDSi      = 0x0103
)

// IDs of languages in NDS banner titles
type TitleLanguage int

const (
	TitleLanguage_Japanese = TitleLanguage(iota)
	TitleLanguage_English
	TitleLanguage_French
	TitleLanguage_German
	TitleLanguage_Italian
	TitleLanguage_Spanish
	TitleLanguage_Chinese
	TitleLanguage_Korean

	// Constant, how many languages are available in the banner
	TitleLanguage_Count
)

var bannerLanguageNames = []string{
	"Japanese",
	"English",
	"French",
	"German",
	"Italian",
	"Spanish",
	"Chinese",
	"Korean",
}

func (t TitleLanguage) String() string {
	if t >= TitleLanguage_Count {
		return "invalid NDS banner language"
	}
	return bannerLanguageNames[t]
}

type banner struct {
	version uint16
	icon    image.PalettedImage
	titles  [8]string
	crcs    [4]uint16
}

// Check if a given title language is valid for banner version.
func (b *banner) checkValidLanguage(language TitleLanguage) error {
	if language >= TitleLanguage_Count {
		return fmt.Errorf("rom title: no language with ID %d exists (max is 7)", language)
	}
	if language == TitleLanguage_Korean && b.version < BannerVersionKorean {
		return fmt.Errorf("rom title: banner version %04X does not support korean", b.version)
	}
	if language == TitleLanguage_Chinese && b.version < BannerVersionChinese {
		return fmt.Errorf("rom title: banner version %04X does not support chinese", b.version)
	}
	return nil
}

// Get binary size of banner based on version
func (b *banner) getSize() int {
	switch b.version {

	// Original + chinese, fits in the same block
	case BannerVersionOriginal, BannerVersionChinese:
		return 0x0A00

	// Has Korean crosses a block
	case BannerVersionKorean:
		return 0x0C00

	// Extended DSi banner
	case BannerVersionDSi:
		return 0x2400

	// Unknown version
	default:
		return -1
	}
}

func SerializeIcon(src image.Image) ([]byte, image.PalettedImage, error) {
	// Validate image size
	bounds := src.Bounds()
	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		return nil, nil, fmt.Errorf("malformed image: bounds.Min.X and bounds.Min.Y should be 0")
	}
	if bounds.Max.X != 32 || bounds.Max.Y != 32 {
		return nil, nil, fmt.Errorf("malformed image: Icon should be exactly 32x32 pixels")
	}

	// Get palette
	palettedDraw := image.NewPaletted(image.Rect(0, 0, 32, 32), nil)
	palMap := make(map[color.RGBA]int)
	palette := make(color.Palette, 1, 16)
	addColor := func(col color.RGBA) (byte, error) {
		if col.A == 0 {
			palette[0] = col
			return 0, nil
		}
		idx, ok := palMap[col]
		if !ok {

			// Transparent pixel
			if col.A != 255 && len(palette) != 0 {
				return 0, fmt.Errorf("malformed image: Pixels cannot have partial transparency")
			}

			// Too many colors
			if len(palette) == 16 {
				return 0, fmt.Errorf("malformed image: Icon can only contain 15 colors + transparency")
			}

			palMap[col] = len(palette)
			palette = append(palette, col)
			palettedDraw.Palette = palette
		}
		return byte(idx), nil
	}
	paletted, isPaletted := src.(image.PalettedImage)
	if isPaletted {
		palette = make(color.Palette, 16)
	}
	palette[0] = color.RGBA{}
	palettedDraw.Palette = palette

	// Create palette and tiles at the same time
	tiles := make([]Tile, 16)
	for i := range tiles {
		t := &tiles[i]
		for j := range 64 {
			x := 8*(i&3) + (j & 7)
			y := 8*(i>>2) + (j >> 3)
			if !isPaletted {
				r, g, b, a := src.At(x, y).RGBA()
				idx, err := addColor(color.RGBA{byte(r), byte(g), byte(b), byte(a)})
				if err != nil {
					return nil, nil, err
				}
				t.Pix[j] = idx
				palettedDraw.SetColorIndex(x, y, idx)
			} else {
				idx := paletted.ColorIndexAt(x, y)
				if idx > 15 {
					return nil, nil, fmt.Errorf("malformed image: Icon can only contain 15 colors + transparency")
				}
				r, g, b, a := src.At(x, y).RGBA()
				if byte(a) != 255 && idx != 0 {
					return nil, nil, fmt.Errorf("malformed image: Pixels cannot have partial transparency")
				}
				palette[idx] = color.RGBA{byte(r), byte(g), byte(b), byte(a)}
				t.Pix[j] = idx
			}
		}
	}

	// Serialize ALL the things!
	tbuf, err := SerializeTiles4BPP(tiles)
	if err != nil {
		return nil, nil, err
	}
	pbuf, err := SerializePalette(palette, 32)
	if err != nil {
		return nil, nil, err
	}
	if !isPaletted {
		paletted = palettedDraw
	}
	return append(tbuf, pbuf...), paletted, nil
}

func OpenBanner(r io.Reader) (*banner, error) {
	b := &banner{}

	// Read data
	tilesRaw := make([]byte, 0x200)
	paletteRaw := make([]byte, 0x20)
	err := ezbin.Read(r,
		&b.version,
		&b.crcs[0],
		&b.crcs[1],
		&b.crcs[2],
		&b.crcs[3],
		ezbin.FillerArray(0x16, byte(0)),
		tilesRaw,
		paletteRaw,
	)
	if err != nil {
		return nil, err
	}

	// Get language count from version + verison verification
	langCount := 0
	switch b.version {
	case BannerVersionOriginal:
		langCount = 6
	case BannerVersionChinese:
		langCount = 7
	case BannerVersionKorean:
		langCount = 8
	case BannerVersionDSi:
		return nil, fmt.Errorf("decode banner: DSi banner mode not yet supported")
	default:
		return nil, fmt.Errorf("decode banner: %04X is not a valid banner version", b.version)
	}

	// Deserialize palette and tiles
	palette := DeserializePalette(paletteRaw, true)
	tiles := DeserializeTiles4BPP(tilesRaw)

	// Turn into one big image
	icon := NewTilemap(4, 4, tiles, palette)
	b.icon = icon
	for i := range icon.Attributes {
		icon.Attributes[i] = TilemapAttributes(i)
	}

	// Read titles
	rawTitle := make([]uint16, 128)
	for i := 0; i < langCount; i++ {
		err := binary.Read(r, binary.LittleEndian, rawTitle)
		if err != nil {
			return nil, err
		}
		str := utf16.Decode(rawTitle)
		b.titles[i] = strings.Trim(string(str), "\x00")
	}

	// Everything worked out :)
	return b, nil
}

func SaveBanner(out io.Writer, b *banner) error {
	buf := make([]byte, 0x2400)
	w := util.NewWriteSeeker(buf)

	// Get language count from version + verison verification
	langCount := 0
	switch b.version {
	case BannerVersionOriginal:
		langCount = 6
	case BannerVersionChinese:
		langCount = 7
	case BannerVersionKorean, BannerVersionDSi:
		langCount = 8
	default:
		return fmt.Errorf("encode banner: %04X is not a valid banner version", b.version)
	}

	// Serialize icon
	icon, _, err := SerializeIcon(b.icon)
	if err != nil {
		return err
	}

	// Serialize titles
	titles := make([]uint16, 0x80*langCount)
	for i, v := range b.titles {
		if i == langCount {
			break
		}
		title := utf16.Encode([]rune(v))
		copy(titles[i*0x80:], title)
	}

	// Finally, write everything
	err = ezbin.Write(w,
		b.version,

		// CRCs, will be filled in later
		uint16(0),
		uint16(0),
		uint16(0),
		uint16(0),

		// Reserved (0-filled)
		ezbin.FillerArray(0x16, byte(0)),

		// Icon tiles and palette
		icon,

		// Titles
		titles,
	)
	if err != nil {
		return err
	}

	// Write padding
	length := w.Pos
	padLength := (0x200 - (length & 0x1FF)) & 0x1FF
	ezbin.Write(w, ezbin.FillerArray(int(padLength), byte(0xFF)))
	dlen := w.Pos

	// Write CRC's
	crcs := []uint16{
		CRC16(buf[0x0020:0x0840]),
		CRC16(buf[0x0020:0x0940]),
		CRC16(buf[0x0020:0x0A40]),
		CRC16(buf[0x0020:0x23C0]),
	}
	w.Seek(0x02, io.SeekStart)
	err = ezbin.Write(w, crcs)
	if err != nil {
		return err
	}

	// Write to output and return
	_, err = out.Write(buf[:dlen])
	return err
}

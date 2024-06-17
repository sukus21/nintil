package nds

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"unicode/utf16"

	"github.com/sukus21/nintil/nds/nitrofs"
	"github.com/sukus21/nintil/util"
	"github.com/sukus21/nintil/util/ezbin"
	"github.com/sukus21/nintil/util/mapping"
)

// Nintendo DS ROM structure.
// Contains most of the things you probably want to get from a ROM file. m
type Rom struct {
	raw        []byte
	reader     io.ReadSeeker
	mapping    *mapping.Mapping
	header     *header
	banner     *banner
	Filesystem nitrofs.NitroFS
	Arm9Binary []byte
	Arm7Binary []byte
}

func (o *Rom) String() string {
	return o.mapping.String()
}

// Open a new ROM.
// TODO: ROM validation.
// TODO: Don't load the enture ROM into memory for no reason
func OpenROM(r io.Reader) (*Rom, error) {
	buf, _ := io.ReadAll(r)
	rom := &Rom{
		reader:  bytes.NewReader(buf),
		raw:     buf,
		mapping: mapping.NewMapping(0x4000000),
	}
	if err := rom.openHeader(); err != nil {
		return nil, err
	}
	if err := rom.openNitroFS(); err != nil {
		return nil, err
	}
	if err := rom.openBanner(); err != nil {
		return nil, err
	}

	rom.Arm9Binary = make([]byte, rom.header.Arm9Size)
	rom.reader.Seek(int64(rom.header.Arm9RomOffset), io.SeekStart)
	if _, err := rom.reader.Read(rom.Arm9Binary); err != nil {
		return nil, err
	}
	rom.mapping.AddAt(mappingNameArm9Binary, rom.header.Arm9RomOffset, rom.header.Arm9Size)

	rom.Arm7Binary = make([]byte, rom.header.Arm7Size)
	rom.reader.Seek(int64(rom.header.Arm7RomOffset), io.SeekStart)
	if _, err := rom.reader.Read(rom.Arm7Binary); err != nil {
		return nil, err
	}
	rom.mapping.AddAt(mappingNameArm7Binary, rom.header.Arm7RomOffset, rom.header.Arm7Size)

	return rom, nil
}

// Serialize ROM.
// TODO: ROM validation.
func SaveROM(o *Rom, out io.Writer) error {

	// TODO: currently only worried about Partners in Time PAL, which is always 64MB.
	romSize := 0x4000000
	wRaw := util.NewWriteSeeker(make([]byte, romSize))
	w := util.NewWriteAtSeeker(wRaw)
	m := mapping.NewMapping(uint32(romSize))
	h := *o.header
	nh := &h

	// Write ARM9 binary
	pos, _ := w.Seek(0x4000, io.SeekStart)
	if err := ezbin.WritePadded(w, 0x0200, 0xFF, o.Arm9Binary); err != nil {
		return err
	}
	nh.Arm9RomOffset = uint32(pos)
	nh.Arm9Size = uint32(len(o.Arm9Binary))
	nh.Arm9ExecuteAddress = nh.Arm9Destination + 0x0800

	// Write ARM7 binary
	pos, _ = w.Seek(0, io.SeekCurrent)
	if err := ezbin.WritePadded(w, 0x0200, 0xFF, o.Arm7Binary); err != nil {
		return err
	}
	nh.Arm7RomOffset = uint32(pos)
	nh.Arm7Size = uint32(len(o.Arm7Binary))
	nh.Arm7ExecuteAddress = nh.Arm7Destination

	// Serialize banner
	pos, _ = w.Seek(0, io.SeekCurrent)
	if err := SaveBanner(w, o.banner); err != nil {
		return err
	}
	nh.BannerOffset = uint32(pos)

	// Serialize NitroFS
	ezbin.Align(w, 0x0200)
	nfsInfo, err := nitrofs.Build(w, o.Filesystem, m)
	if err != nil {
		return err
	}

	// Update header
	nh.ApplyNitroFSInfo(nfsInfo)
	nh.UpdateChecksum()

	// Finally, serialize header
	w.Seek(0, io.SeekStart)
	if err := SaveHeader(w, nh); err != nil {
		return err
	}

	// Copy all of this to the output writer
	_, err = out.Write(wRaw.Buf)
	return err
}

// Read new header.
// Should only be called once.
func (o *Rom) openHeader() error {
	o.reader.Seek(0, io.SeekStart)
	h, err := OpenHeader(o.reader)
	if err != nil {
		return err
	}

	// Yay :)
	o.header = h
	o.mapping.AddAt(mappingNameHeader, 0x00, 0x4000)
	return err
}

// Read the in-ROM filesystem.
func (o *Rom) openNitroFS() error {
	fs := nitrofs.FromROM(o.reader, o.header.GetNitroFSInfo(), o.mapping)
	o.Filesystem = fs
	return nil
}

// Read the ROM's banner (titles + icon).
// TODO: does not work with DSi animated icons.
func (o *Rom) openBanner() error {
	o.reader.Seek(int64(o.header.BannerOffset), io.SeekStart)
	b, err := OpenBanner(o.reader)
	if err != nil {
		return err
	}
	_, err = o.mapping.AddAt(mappingBanner, o.header.BannerOffset, uint32(b.getSize()))
	if err != nil {
		return err
	}
	o.banner = b
	return nil
}

// Expose header to outside world.
func (o *Rom) GetHeader() *header {
	return o.header
}

// Get current ROM icon.
func (o *Rom) GetIcon() image.PalettedImage {
	return o.banner.icon
}

// Set the ROM icon (if possible).
func (o *Rom) SetIcon(src image.Image) error {
	_, paletted, err := SerializeIcon(src)
	if err != nil {
		return err
	}

	// Inject into ROM
	o.banner.icon = paletted
	return nil
}

// Get title in specified language.
func (o *Rom) GetTitle(language TitleLanguage) (string, error) {
	if err := o.banner.checkValidLanguage(language); err != nil {
		return "", err
	}
	return o.banner.titles[language], nil
}

// Set title in specified language.
// TODO: validate newlines.
func (o *Rom) SetTitle(title string, language TitleLanguage) error {
	if err := o.banner.checkValidLanguage(language); err != nil {
		return err
	}
	enc := append(utf16.Encode([]rune(title)), 0x0000)
	if len(enc) > 128 {
		return fmt.Errorf("set ROM title: title too long (max is 127 encoded chars, got %d)", len(enc))
	}
	o.banner.titles[language] = title
	return nil
}

func (o *Rom) GetBannerVersion() uint16 {
	return o.banner.version
}

func (o *Rom) SetBannerVersion(version uint16) {
	o.banner.version = version
}

func (o *Rom) WhatsHere(at uint32) *mapping.MappingEntry {
	return o.mapping.Find(at)
}

func CRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i])
		for j := uint32(0); j < 8; j++ {
			o := crc
			crc = (crc >> 1)
			if o&1 != 0 {
				crc ^= 0xA001
			}
		}
	}
	return crc
}

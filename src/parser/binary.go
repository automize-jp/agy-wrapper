package parser

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type Segment struct {
	FileOff uint64
	VAddr   uint64
	Size    uint64
}

type PESection struct {
	FileOff         uint64
	VAddr           uint64
	Size            uint64
	Name            string
	Characteristics uint32
}

type RodataRegion struct {
	FileOff uint64
	Size    uint64
	VAddr   uint64
}

type PatchTarget struct {
	Kind   string
	Offset uint64
}

type BinaryInfo struct {
	Format        string // "pe", "elf", "macho", "macho32"
	Segments      []Segment
	SegmentFlags  map[uint64]uint32
	RodataRegions []RodataRegion
	PESections    []PESection
	RelaDynOff    uint64
	RelaDynSz     uint64
	ImageBase     uint64
	Data          []byte
}

func ParseBinary(d []byte) (*BinaryInfo, error) {
	info := &BinaryInfo{Data: d}
	if len(d) >= 4 && bytes.Equal(d[:4], []byte("\x7fELF")) {
		info.Format = "elf"
		info.parseELF()
	} else if len(d) >= 4 && bytes.Equal(d[:4], []byte("\xcf\xfa\xed\xfe")) {
		info.Format = "macho"
		info.parseMachO()
	} else if len(d) >= 4 && bytes.Equal(d[:4], []byte("\xce\xfa\xed\xfe")) {
		info.Format = "macho32"
		info.parseMachO()
	} else if len(d) > 0x40 && bytes.Equal(d[:2], []byte("MZ")) {
		peOff := binary.LittleEndian.Uint32(d[0x3C : 0x3C+4])
		if int(peOff+4) <= len(d) && bytes.Equal(d[peOff:peOff+4], []byte("PE\x00\x00")) {
			info.Format = "pe"
			info.parsePE(peOff)
		}
	}
	if info.Format == "" {
		return nil, errors.New("unknown binary format")
	}
	return info, nil
}

func (info *BinaryInfo) FileToVA(foff uint64) (uint64, bool) {
	for _, seg := range info.Segments {
		if foff >= seg.FileOff && foff < seg.FileOff+seg.Size {
			return seg.VAddr + (foff - seg.FileOff), true
		}
	}
	return 0, false
}

func (info *BinaryInfo) VAToFile(va uint64) (uint64, bool) {
	for _, seg := range info.Segments {
		if va >= seg.VAddr && va < seg.VAddr+seg.Size {
			return seg.FileOff + (va - seg.VAddr), true
		}
	}
	return 0, false
}

func (info *BinaryInfo) Platform() string {
	switch info.Format {
	case "pe":
		return "windows"
	case "elf":
		return "linux"
	case "macho", "macho32":
		return "macos"
	}
	return "unknown"
}

func (info *BinaryInfo) FindStringRef(promptVA uint64, expectedLen uint64) ([]PatchTarget, uint64, error) {
	if info.Format == "pe" {
		return info.findRefPE(promptVA, expectedLen)
	} else if info.Format == "elf" && info.RelaDynSz > 0 {
		return info.findRefELF(promptVA, expectedLen)
	} else if info.Format == "macho" || info.Format == "macho32" {
		return info.findRefMachO(promptVA, expectedLen)
	}
	return nil, 0, errors.New("unsupported binary format or missing relocation metadata")
}

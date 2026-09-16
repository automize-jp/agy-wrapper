package parser

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func (info *BinaryInfo) parsePE(peOff uint32) {
	d := info.Data
	coffOff := peOff + 4
	numSections := binary.LittleEndian.Uint16(d[coffOff+2 : coffOff+4])
	optHdrSize := binary.LittleEndian.Uint16(d[coffOff+16 : coffOff+18])

	optOff := coffOff + 20
	optMagic := binary.LittleEndian.Uint16(d[optOff : optOff+2])

	if optMagic == 0x20b { // PE32+ (64-bit)
		info.ImageBase = binary.LittleEndian.Uint64(d[optOff+24 : optOff+32])
	} else if optMagic == 0x10b { // PE32 (32-bit)
		info.ImageBase = uint64(binary.LittleEndian.Uint32(d[optOff+28 : optOff+32]))
	} else {
		info.Format = ""
		return
	}

	sectionOff := optOff + uint32(optHdrSize)
	for i := 0; i < int(numSections); i++ {
		off := sectionOff + uint32(i*40)
		if int(off+40) > len(d) {
			break
		}
		rawName := d[off : off+8]
		name := string(bytes.TrimRight(rawName, "\x00"))
		virtualAddress := binary.LittleEndian.Uint32(d[off+12 : off+16])
		rawSize := binary.LittleEndian.Uint32(d[off+16 : off+20])
		rawOffset := binary.LittleEndian.Uint32(d[off+20 : off+24])
		characteristics := binary.LittleEndian.Uint32(d[off+36 : off+40])

		if rawSize > 0 {
			vaddr := info.ImageBase + uint64(virtualAddress)
			info.Segments = append(info.Segments, Segment{
				FileOff: uint64(rawOffset),
				VAddr:   vaddr,
				Size:    uint64(rawSize),
			})
			info.PESections = append(info.PESections, PESection{
				FileOff:         uint64(rawOffset),
				VAddr:           vaddr,
				Size:            uint64(rawSize),
				Name:            name,
				Characteristics: characteristics,
			})
		}
	}
}

func (info *BinaryInfo) findRefPE(promptVA uint64, expectedLen uint64) ([]PatchTarget, uint64, error) {
	vaBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(vaBytes, promptVA)
	d := info.Data

	for _, sec := range info.PESections {
		if sec.Characteristics&0x20000000 != 0 { // IMAGE_SCN_MEM_EXECUTE
			continue
		}
		segStart := int(sec.FileOff)
		segEnd := int(sec.FileOff + sec.Size)
		if segEnd > len(d) {
			segEnd = len(d)
		}

		idx := segStart
		for idx < segEnd {
			pos := bytes.Index(d[idx:segEnd], vaBytes)
			if pos < 0 {
				break
			}
			absPos := idx + pos
			if absPos+16 <= len(d) {
				oldLen := binary.LittleEndian.Uint64(d[absPos+8 : absPos+16])
				diff := int64(oldLen) - int64(expectedLen)
				if diff < 0 {
					diff = -diff
				}
				maxDiff := int64(expectedLen)
				if maxDiff < 5000 {
					maxDiff = 5000
				}
				if diff < maxDiff {
					targets := []PatchTarget{
						{Kind: "ptr_value", Offset: uint64(absPos)},
						{Kind: "string_len", Offset: uint64(absPos + 8)},
					}
					return targets, oldLen, nil
				}
			}
			idx = absPos + 1
		}
	}
	return nil, 0, errors.New("string reference not found in PE sections")
}

package parser

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func (info *BinaryInfo) parseMachO() {
	d := info.Data
	if len(d) < 32 {
		return
	}
	ncmds := binary.LittleEndian.Uint32(d[16:20])
	off := uint32(32)
	for i := uint32(0); i < ncmds; i++ {
		if int(off+8) > len(d) {
			break
		}
		cmd := binary.LittleEndian.Uint32(d[off : off+4])
		cmdsize := binary.LittleEndian.Uint32(d[off+4 : off+8])
		if cmd == 0x19 { // LC_SEGMENT_64
			if int(off+64) <= len(d) {
				vmaddr := binary.LittleEndian.Uint64(d[off+24 : off+32])
				fileoff := binary.LittleEndian.Uint64(d[off+40 : off+48])
				filesize := binary.LittleEndian.Uint64(d[off+48 : off+56])
				initprot := binary.LittleEndian.Uint32(d[off+60 : off+64])
				info.Segments = append(info.Segments, Segment{
					FileOff: fileoff,
					VAddr:   vmaddr,
					Size:    filesize,
				})
				if info.SegmentFlags == nil {
					info.SegmentFlags = make(map[uint64]uint32)
				}
				info.SegmentFlags[fileoff] = initprot
				if (initprot&2) == 0 && (initprot&4) == 0 && filesize >= 128 {
					info.RodataRegions = append(info.RodataRegions, RodataRegion{
						FileOff: fileoff,
						Size:    filesize,
						VAddr:   vmaddr,
					})
				}
			}
		}
		off += cmdsize
	}
}

func (info *BinaryInfo) findRefMachO(promptVA uint64, expectedLen uint64) ([]PatchTarget, uint64, error) {
	vaBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(vaBytes, promptVA)
	d := info.Data

	for _, seg := range info.Segments {
		if seg.FileOff == 0 || seg.Size < 1000000 {
			continue
		}
		segStart := int(seg.FileOff)
		segEnd := int(seg.FileOff + seg.Size)
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
	return nil, 0, errors.New("string reference not found in Mach-O __DATA")
}

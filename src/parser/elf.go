package parser

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func (info *BinaryInfo) parseELF() {
	d := info.Data
	if len(d) < 0x40 {
		return
	}
	e_phoff := binary.LittleEndian.Uint64(d[0x20:0x28])
	e_phentsize := binary.LittleEndian.Uint16(d[0x36:0x38])
	e_phnum := binary.LittleEndian.Uint16(d[0x38:0x3a])

	for i := 0; i < int(e_phnum); i++ {
		off := e_phoff + uint64(i)*uint64(e_phentsize)
		if int(off+36) > len(d) {
			break
		}
		p_type := binary.LittleEndian.Uint32(d[off : off+4])
		if p_type == 1 { // PT_LOAD
			p_offset := binary.LittleEndian.Uint64(d[off+8 : off+16])
			p_vaddr := binary.LittleEndian.Uint64(d[off+16 : off+24])
			p_filesz := binary.LittleEndian.Uint64(d[off+32 : off+40])
			info.Segments = append(info.Segments, Segment{
				FileOff: p_offset,
				VAddr:   p_vaddr,
				Size:    p_filesz,
			})
		}
	}

	e_shoff := binary.LittleEndian.Uint64(d[0x28:0x30])
	e_shentsize := binary.LittleEndian.Uint16(d[0x3a:0x3c])
	e_shnum := binary.LittleEndian.Uint16(d[0x3c:0x3e])
	e_shstrndx := binary.LittleEndian.Uint16(d[0x3e:0x40])

	if e_shnum == 0 || e_shstrndx >= e_shnum {
		return
	}

	shstr_hdr := e_shoff + uint64(e_shstrndx)*uint64(e_shentsize)
	if int(shstr_hdr+40) > len(d) {
		return
	}
	shstr_offset := binary.LittleEndian.Uint64(d[shstr_hdr+24 : shstr_hdr+32])
	shstr_size := binary.LittleEndian.Uint64(d[shstr_hdr+32 : shstr_hdr+40])
	if int(shstr_offset+shstr_size) > len(d) {
		return
	}
	shstrtab := d[shstr_offset : shstr_offset+shstr_size]

	for i := 0; i < int(e_shnum); i++ {
		off := e_shoff + uint64(i)*uint64(e_shentsize)
		if int(off+40) > len(d) {
			break
		}
		sh_name_idx := binary.LittleEndian.Uint32(d[off : off+4])
		if int(sh_name_idx) < len(shstrtab) {
			end := bytes.IndexByte(shstrtab[sh_name_idx:], 0x00)
			if end != -1 {
				name := string(shstrtab[sh_name_idx : int(sh_name_idx)+end])
				if name == ".rela.dyn" {
					info.RelaDynOff = binary.LittleEndian.Uint64(d[off+24 : off+32])
					info.RelaDynSz = binary.LittleEndian.Uint64(d[off+32 : off+40])
					break
				}
			}
		}
	}
}

func (info *BinaryInfo) findRefELF(promptVA uint64, expectedLen uint64) ([]PatchTarget, uint64, error) {
	d := info.Data
	n := int(info.RelaDynSz / 24)
	for i := 0; i < n; i++ {
		off := int(info.RelaDynOff) + i*24
		if off+24 > len(d) {
			break
		}
		rInfo := binary.LittleEndian.Uint64(d[off+8 : off+16])
		rType := uint32(rInfo & 0xFFFFFFFF)
		if rType != 8 { // R_X86_64_RELATIVE
			continue
		}
		rAddend := binary.LittleEndian.Uint64(d[off+16 : off+24])
		if rAddend != promptVA {
			continue
		}
		rOffsetVA := binary.LittleEndian.Uint64(d[off : off+8])
		foff, ok := info.VAToFile(rOffsetVA)
		if !ok || int(foff+16) > len(d) {
			continue
		}
		oldLen := binary.LittleEndian.Uint64(d[foff+8 : foff+16])
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
				{Kind: "rela_addend", Offset: uint64(off + 16)},
				{Kind: "string_len", Offset: foff + 8},
			}
			return targets, oldLen, nil
		}
	}
	return nil, 0, errors.New("string reference not found in .rela.dyn")
}

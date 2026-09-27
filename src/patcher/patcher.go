package patcher

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"agy-wrapper/src/parser"
	"agy-wrapper/src/prompt"
)

func GetBinaryVersion(binPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func ValidateBackup(realPath, origPath string, realData, origData, stored []byte) (bool, string) {
	// 1. Size check
	if len(origData) != len(realData) {
		return false, fmt.Sprintf("size mismatch: backup is %d bytes, target binary is %d bytes (different agy build)", len(origData), len(realData))
	}

	// 2. Pristine prompt check
	_, _, origExact := FindPrompt(origData, stored)
	if !origExact {
		return false, "backup binary does not contain the pristine unpatched prompt (file may be corrupted or already patched)"
	}

	// 3. Executable version check
	realVer, errReal := GetBinaryVersion(realPath)
	origVer, errOrig := GetBinaryVersion(origPath)
	if errReal == nil && errOrig == nil && realVer != "" && origVer != "" {
		if realVer != origVer {
			return false, fmt.Sprintf("version mismatch: backup version '%s' does not match target binary version '%s'", origVer, realVer)
		}
	}

	return true, ""
}

func FindNullRegion(data []byte, minSize int) (int, int, bool) {
	probeLen := minSize
	if probeLen > 4096 {
		probeLen = 4096
	}
	probe := make([]byte, probeLen)

	bestStart := -1
	bestEnd := -1
	bestSize := 0

	i := 0
	for i < len(data) {
		pos := bytes.Index(data[i:], probe)
		if pos < 0 {
			break
		}
		absPos := i + pos
		start := absPos
		for start > 0 && data[start-1] == 0 {
			start--
		}
		end := absPos + len(probe)
		for end < len(data) && data[end] == 0 {
			end++
		}
		sz := end - start
		if sz >= minSize && sz > bestSize {
			bestStart = start
			bestEnd = end
			bestSize = sz
		}
		i = end
	}

	if bestStart != -1 {
		return bestStart, bestEnd, true
	}
	return 0, 0, false
}

func FindPrompt(data []byte, stored []byte) (int, int, bool) {
	idx := bytes.Index(data, stored)
	if idx >= 0 {
		return idx, idx + len(stored), true
	}

	// Strategy: XML tags
	openTag := []byte("<GRILL_ME>")
	closeTag := []byte("</GRILL_ME>")
	s := bytes.Index(data, openTag)
	if s >= 0 {
		e := bytes.Index(data[s:], closeTag)
		if e >= 0 {
			return s, s + e + len(closeTag), false
		}
	}

	// Progressive line matching
	lines := bytes.Split(stored, []byte("\n"))
	var filteredLines [][]byte
	for _, l := range lines {
		if len(bytes.TrimSpace(l)) > 0 {
			filteredLines = append(filteredLines, l)
		}
	}
	if len(filteredLines) == 0 {
		return -1, -1, false
	}

	firstLine := filteredLines[0]
	var positions []int
	searchPos := 0
	for {
		p := bytes.Index(data[searchPos:], firstLine)
		if p < 0 {
			break
		}
		positions = append(positions, searchPos+p)
		searchPos += p + 1
	}

	for i := 1; i < len(filteredLines) && len(positions) > 1; i++ {
		prefix := bytes.Join(filteredLines[:i+1], []byte("\n"))
		var newPositions []int
		for _, p := range positions {
			if p+len(prefix) <= len(data) && bytes.Equal(data[p:p+len(prefix)], prefix) {
				newPositions = append(newPositions, p)
			}
		}
		positions = newPositions
	}

	if len(positions) == 1 {
		start := positions[0]
		end := start
		limit := start + 200000
		if limit > len(data) {
			limit = len(data)
		}
		for end < limit && (data[end] >= 0x20 || data[end] == '\n' || data[end] == '\r' || data[end] == '\t') {
			end++
		}
		return start, end, false
	}

	return -1, -1, false
}

func ApplyContent(buf []byte, info *parser.BinaryInfo, start int, end int, newContent []byte, dryRun bool) (string, error) {
	oldLen := end - start
	newLen := len(newContent)
	delta := newLen - oldLen

	var logMsg strings.Builder
	logMsg.WriteString(fmt.Sprintf("  Old: %d bytes, New: %d bytes (%+d)\n", oldLen, newLen, delta))

	if delta <= 0 {
		if !dryRun {
			copy(buf[start:], newContent)
			for i := start + newLen; i < end; i++ {
				buf[i] = ' '
			}
		}
		logMsg.WriteString("  In-place ✓\n")
		return logMsg.String(), nil
	}

	logMsg.WriteString(fmt.Sprintf("  Content grew by %d bytes, searching for string reference...\n", delta))
	promptVA, ok := info.FileToVA(uint64(start))
	if !ok {
		return logMsg.String(), fmt.Errorf("cannot compute VA for offset 0x%x", start)
	}

	targets, oldRefLen, err := info.FindStringRef(promptVA, uint64(oldLen))
	if err != nil {
		return logMsg.String(), fmt.Errorf("string ref not found (VA=0x%x): %w", promptVA, err)
	}

	logMsg.WriteString(fmt.Sprintf("  String ref found: len=%d\n", oldRefLen))
	for _, t := range targets {
		logMsg.WriteString(fmt.Sprintf("    %s at file offset 0x%x\n", t.Kind, t.Offset))
	}

	nullStart, nullEnd, found := FindNullRegion(buf, newLen+64)
	if !found {
		return logMsg.String(), errors.New("no null region large enough")
	}

	newVA, ok := info.FileToVA(uint64(nullStart))
	if !ok {
		return logMsg.String(), errors.New("null region not in any loadable segment")
	}

	logMsg.WriteString(fmt.Sprintf("  Relocating to offset 0x%x (VA 0x%x), %d bytes available\n", nullStart, newVA, nullEnd-nullStart))

	if !dryRun {
		copy(buf[nullStart:], newContent)
		for _, t := range targets {
			if strings.Contains(t.Kind, "addend") || strings.Contains(t.Kind, "ptr") {
				binary.LittleEndian.PutUint64(buf[t.Offset:t.Offset+8], newVA)
			} else if strings.Contains(t.Kind, "len") {
				binary.LittleEndian.PutUint64(buf[t.Offset:t.Offset+8], uint64(newLen))
			}
		}
		for i := start; i < end; i++ {
			buf[i] = 0
		}
	}

	logMsg.WriteString(fmt.Sprintf("  Relocated %d bytes ✓\n", newLen))
	return logMsg.String(), nil
}

// PatchBinary unifies checking, dry-run simulation, and applying the prompt patches to the binary
func PatchBinary(binaryPath string, dryRun bool) (string, error) {
	var out strings.Builder
	p := func(format string, a ...any) {
		out.WriteString(fmt.Sprintf(format+"\n", a...))
	}

	realPath, err := filepath.EvalSymlinks(binaryPath)
	if err != nil {
		realPath = binaryPath
	}
	originalPath := realPath + ".original"

	p("════════════════════════════════════════════════════════════")
	p("Binary: %s", realPath)

	data, err := os.ReadFile(realPath)
	if err != nil {
		return out.String(), fmt.Errorf("failed to read binary from %s: %w", realPath, err)
	}

	buf := make([]byte, len(data))
	copy(buf, data)

	p("Size: %.1f MB", float64(len(buf))/1e6)

	if len(buf) < 20*1024*1024 {
		return out.String(), fmt.Errorf("refusing to patch binary smaller than 20MB (target appears to be a wrapper or non-antigravity binary)")
	}

	info, err := parser.ParseBinary(buf)
	if err != nil {
		return out.String(), fmt.Errorf("failed to parse binary: %w", err)
	}

	platform := info.Platform()
	p("Format: %s, %d segments", strings.ToUpper(info.Format), len(info.Segments))
	p("Platform: %s", platform)

	tools := []string{"grill-me"}
	p("Auto-discovered tools to patch: %s", strings.Join(tools, ", "))

	for _, toolName := range tools {
		p("\n[%s] Processing tool...", toolName)

		stored, err := prompt.LoadPromptFile(platform, toolName, "")
		if err != nil {
			return out.String(), err
		}

		fixed, err := prompt.LoadPromptFile(platform, toolName, ".fixed")
		if err != nil {
			return out.String(), err
		}

		start, end, exact := FindPrompt(buf, stored)
		if start == -1 {
			return out.String(), fmt.Errorf("[%s] Prompt not found in binary (possibly modified or corrupted)", toolName)
		}

		extracted := buf[start:end]
		h := sha256.Sum256(extracted)
		extractedSum := hex.EncodeToString(h[:])

		p("[%s] Found %d bytes at offset 0x%x", toolName, len(extracted), start)
		p("[%s] SHA256: %s", toolName, extractedSum)

		if exact {
			// 1. Binary is unpatched (clean upstream binary)
			p("[%s] Target binary is unpatched (clean upstream binary).", toolName)
			if !dryRun {
				// Force create / overwrite backup from pristine data
				p("[%s] Force creating backup → %s", toolName, originalPath)
				if err := os.WriteFile(originalPath, data, 0755); err != nil {
					return out.String(), fmt.Errorf("failed to create backup %s: %w", originalPath, err)
				}
			}
		} else {
			// 2. Binary is patched (exact == false)
			fixedTrimmed := bytes.TrimRight(fixed, "\r\n")
			extractedTrimmed := bytes.TrimRight(extracted, "\r\n")
			if bytes.Equal(extractedTrimmed, fixedTrimmed) {
				p("[%s] Binary is already patched with target fixed prompt ✓", toolName)
				continue
			}

			// Binary has an OLD patch
			p("[%s] Old patch detected in target binary (prompt differs from current fixed target).", toolName)

			// Check if .original backup exists
			if _, errStat := os.Stat(originalPath); os.IsNotExist(errStat) {
				return out.String(), fmt.Errorf("[%s] Target binary '%s' contains an old patch, but backup '%s' does not exist.\nCannot safely restore and re-patch.\nTo recover:\n  1. Run update directly on target binary: %s update\n  2. Or download a fresh upstream binary to replace %s", toolName, realPath, originalPath, realPath, realPath)
			}

			// Read and validate backup
			origData, errOrig := os.ReadFile(originalPath)
			if errOrig != nil {
				return out.String(), fmt.Errorf("[%s] Failed to read backup %s: %w", toolName, originalPath, errOrig)
			}

			ok, reason := ValidateBackup(realPath, originalPath, data, origData, stored)
			if !ok {
				return out.String(), fmt.Errorf("[%s] Backup '%s' is invalid (%s).\nCannot safely restore and re-patch.\nTo recover:\n  1. Run update directly on target binary: %s update\n  2. Or download a fresh upstream binary to replace %s", toolName, originalPath, reason, realPath, realPath)
			}

			p("[%s] Valid backup verified (%s). Restoring and re-basing patch...", toolName, originalPath)
			buf = make([]byte, len(origData))
			copy(buf, origData)
			data = origData
			info, _ = parser.ParseBinary(buf)
			origStart, origEnd, origExact := FindPrompt(buf, stored)
			if !origExact {
				return out.String(), fmt.Errorf("[%s] Internal error: backup prompt search failed after validation", toolName)
			}
			start = origStart
			end = origEnd
		}

		p("[%s] Checksum OK ✓", toolName)

		logMsg, err := ApplyContent(buf, info, start, end, fixed, dryRun)
		out.WriteString(logMsg)
		if err != nil {
			return out.String(), fmt.Errorf("[%s] Failed to apply patch: %w", toolName, err)
		}
	}

	if dryRun {
		p("\nDRY RUN — no changes written.")
		return out.String(), nil
	}

	if bytes.Equal(buf, data) {
		p("\nBinary is already up to date. No changes written.")
		return out.String(), nil
	}

	// Backup original binary if not exists yet
	if _, err := os.Stat(originalPath); os.IsNotExist(err) {
		p("Backup → %s", originalPath)
		_ = os.WriteFile(originalPath, data, 0755)
	}

	// Write patched binary
	p("\nWriting %s...", realPath)
	tmpPath := realPath + ".patching"
	if err := os.WriteFile(tmpPath, buf, 0755); err != nil {
		return out.String(), fmt.Errorf("failed to write patched binary to %s: %w", tmpPath, err)
	}

	// 2-phase swap: Windows allows renaming a running executable to .old
	oldPath := fmt.Sprintf("%s.%d.old", realPath, os.Getpid())
	_ = os.Remove(oldPath)

	if err := os.Rename(realPath, oldPath); err != nil {
		if errDirect := os.Rename(tmpPath, realPath); errDirect != nil {
			_ = os.Remove(tmpPath)
			return out.String(), fmt.Errorf("failed to replace %s: %w", realPath, errDirect)
		}
	} else {
		if err := os.Rename(tmpPath, realPath); err != nil {
			_ = os.Rename(oldPath, realPath) // rollback
			_ = os.Remove(tmpPath)
			return out.String(), fmt.Errorf("failed to move patched binary into %s: %w", realPath, err)
		}
		_ = os.Remove(oldPath)
	}
	p("Done ✓")

	return out.String(), nil
}

package meta

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"agy-wrapper/src/patcher"
)

type PatchMeta struct {
	RealBin string `json:"real_bin"`
	Size    int64  `json:"size"`
	Mtime   int64  `json:"mtime"`
	SHA256  string `json:"sha256"`
}

func GetMetaFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	geminiDir := filepath.Join(home, ".gemini")
	_ = os.MkdirAll(geminiDir, 0755)
	return filepath.Join(geminiDir, ".latest-patched-version")
}

func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func EnsurePatchedSilently(realBin string) {
	fi, err := os.Stat(realBin)
	if err != nil {
		return
	}

	metaFile := GetMetaFile()
	var m PatchMeta

	// 1. Fast Path (0ms): If size and mtime match -> Binary has not changed
	if data, err := os.ReadFile(metaFile); err == nil {
		_ = json.Unmarshal(data, &m)
		if m.RealBin == realBin && m.Size == fi.Size() && m.Mtime == fi.ModTime().Unix() {
			return
		}
	}

	// 2. If size or mtime changed: Compute SHA256 of current file to verify
	currentSHA, err := FileSHA256(realBin)
	if err == nil && m.RealBin == realBin && m.SHA256 != "" && currentSHA == m.SHA256 {
		// SHA256 matches cached patched record -> Already patched!
		m.Mtime = fi.ModTime().Unix()
		m.Size = fi.Size()
		if data, err := json.Marshal(m); err == nil {
			_ = os.WriteFile(metaFile, data, 0644)
		}
		return
	}

	// 3. SHA256 or size differ: Upstream binary changed -> Proceed to patch
	out, err := patcher.PatchBinary(realBin, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, out)
		fmt.Fprintf(os.Stderr, "[AGY Patch Error] %v\n", err)
		os.Exit(1)
	}

	// 4. Successful patch: Cache new SHA256, size, and mod-time
	if fiNew, err := os.Stat(realBin); err == nil {
		newSHA, _ := FileSHA256(realBin)
		newMeta := PatchMeta{
			RealBin: realBin,
			Size:    fiNew.Size(),
			Mtime:   fiNew.ModTime().Unix(),
			SHA256:  newSHA,
		}
		if data, err := json.Marshal(newMeta); err == nil {
			_ = os.WriteFile(metaFile, data, 0644)
		}
	}
}

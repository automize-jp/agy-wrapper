package meta

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"agy-wrapper/src/patcher"
	"agy-wrapper/src/prompt"
)

type PatchMeta struct {
	RealBin string            `json:"real_bin"`
	Size    int64             `json:"size"`
	Mtime   int64             `json:"mtime"`
	SHA256  string            `json:"sha256"`
	Prompts map[string]string `json:"prompts,omitempty"`
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

func RuntimePlatform() string {
	if runtime.GOOS == "darwin" {
		return "macos"
	}
	return runtime.GOOS
}

func GetTargetPromptsSHA256(tools []string) map[string]string {
	res := make(map[string]string)
	platform := RuntimePlatform()
	for _, tool := range tools {
		if fixed, err := prompt.LoadPromptFile(platform, tool, ".fixed"); err == nil {
			h := sha256.Sum256(fixed)
			res[tool] = hex.EncodeToString(h[:])
		}
	}
	return res
}

func PromptsMatch(stored map[string]string, current map[string]string) bool {
	if len(current) == 0 {
		return true
	}
	if len(stored) == 0 {
		return false
	}
	for k, v := range current {
		if stored[k] != v {
			return false
		}
	}
	return true
}

func EnsurePatchedSilently(realBin string) {
	fi, err := os.Stat(realBin)
	if err != nil {
		return
	}

	metaFile := GetMetaFile()
	var m PatchMeta
	metaReadOk := false
	if data, err := os.ReadFile(metaFile); err == nil {
		if err := json.Unmarshal(data, &m); err == nil {
			metaReadOk = true
		}
	}

	tools := []string{"grill-me"}
	targetPrompts := GetTargetPromptsSHA256(tools)

	// 1. Fast Path (0ms): If size, mtime match AND target prompts match -> Binary has not changed
	if metaReadOk && m.RealBin == realBin && m.Size == fi.Size() && m.Mtime == fi.ModTime().Unix() {
		if PromptsMatch(m.Prompts, targetPrompts) {
			return
		}
	}

	// 2. If size or mtime changed: Compute SHA256 of current file to verify
	currentSHA, err := FileSHA256(realBin)
	if err == nil && metaReadOk && m.RealBin == realBin && m.SHA256 != "" && currentSHA == m.SHA256 {
		if PromptsMatch(m.Prompts, targetPrompts) {
			// SHA256 matches cached patched record & prompts match -> Already patched!
			m.Mtime = fi.ModTime().Unix()
			m.Size = fi.Size()
			if data, err := json.Marshal(m); err == nil {
				_ = os.WriteFile(metaFile, data, 0644)
			}
			return
		}
	}

	// 3. SHA256 differ or prompts changed -> Proceed to patch
	out, err := patcher.PatchBinary(realBin, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, out)
		fmt.Fprintf(os.Stderr, "[AGY Patch Error] %v\n", err)
		os.Exit(1)
	}

	// 4. Successful patch: Cache new SHA256, size, mod-time, and target prompts
	if fiNew, err := os.Stat(realBin); err == nil {
		newSHA, _ := FileSHA256(realBin)
		newMeta := PatchMeta{
			RealBin: realBin,
			Size:    fiNew.Size(),
			Mtime:   fiNew.ModTime().Unix(),
			SHA256:  newSHA,
			Prompts: targetPrompts,
		}
		if data, err := json.Marshal(newMeta); err == nil {
			_ = os.WriteFile(metaFile, data, 0644)
		}
	}
}

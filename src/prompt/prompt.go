package prompt

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"agy-wrapper/src/parser"
)

//go:embed prompts/*/*
var embeddedPrompts embed.FS

type PromptToolInfo struct {
	Name     string
	OpenTag  string
	CloseTag string
}

var ExtractableTools = []PromptToolInfo{
	{Name: "grill-me", OpenTag: "<GRILL_ME>", CloseTag: "</GRILL_ME>"},
	{Name: "plan", OpenTag: "<PLAN>", CloseTag: "</PLAN>"},
	{Name: "learn", OpenTag: "<LEARN>", CloseTag: "</LEARN>"},
}

// LoadPromptFile reads a prompt file from disk or falls back to embedded assets
func LoadPromptFile(platform, toolName, suffix string) ([]byte, error) {
	fileName := toolName + suffix + ".md"

	// 1. Try disk paths: ./prompts/<platform>/, ./src/prompt/prompts/<platform>/, <exeDir>/prompts/<platform>/
	candidates := []string{
		filepath.Join("prompts", platform, fileName),
		filepath.Join("src", "prompt", "prompts", platform, fileName),
	}
	if selfExe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(selfExe), "prompts", platform, fileName))
	}

	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil {
			return bytes.TrimRight(data, "\r\n"), nil
		}
	}

	// 2. Fallback to embedded prompts
	embedPath := fmt.Sprintf("prompts/%s/%s", platform, fileName)
	if data, err := fs.ReadFile(embeddedPrompts, embedPath); err == nil {
		return bytes.TrimRight(data, "\r\n"), nil
	}

	return nil, fmt.Errorf("prompt file not found: %s", fileName)
}

// ExtractPrompts scans the binary and exports matching prompt XML tags to disk or stdout
func ExtractPrompts(binaryPath string, outDir string, toStdout bool, filterTool string) (string, error) {
	var out strings.Builder
	p := func(format string, a ...any) {
		out.WriteString(fmt.Sprintf(format+"\n", a...))
	}

	realPath, err := filepath.EvalSymlinks(binaryPath)
	if err != nil {
		realPath = binaryPath
	}

	data, err := os.ReadFile(realPath)
	if err != nil {
		return out.String(), fmt.Errorf("failed to read binary from %s: %w", realPath, err)
	}

	info, err := parser.ParseBinary(data)
	if err != nil {
		return out.String(), fmt.Errorf("failed to parse binary: %w", err)
	}

	platform := info.Platform()
	if outDir == "" {
		outDir = filepath.Join("prompts", platform)
	}

	if !toStdout {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return out.String(), fmt.Errorf("failed to create output directory %s: %w", outDir, err)
		}
		p("════════════════════════════════════════════════════════════")
		p("Binary: %s", realPath)
		p("Platform: %s", platform)
		if filterTool != "" {
			p("Filter Tool: %s", filterTool)
		}
		p("Export Directory: %s", outDir)
		p("════════════════════════════════════════════════════════════")
	}

	// Check if backup original binary exists
	originalPath := realPath + ".original"
	origData, _ := os.ReadFile(originalPath)

	extractedCount := 0
	for _, tool := range ExtractableTools {
		if filterTool != "" && !strings.EqualFold(tool.Name, filterTool) {
			continue
		}
		openTag := []byte(tool.OpenTag)
		closeTag := []byte(tool.CloseTag)

		s := bytes.Index(data, openTag)
		if s < 0 {
			if !toStdout {
				p("[-] %s (%s): Tag not found in binary", tool.Name, tool.OpenTag)
			}
			continue
		}

		e := bytes.Index(data[s:], closeTag)
		if e < 0 {
			if !toStdout {
				p("[-] %s (%s): Closing tag %s not found", tool.Name, tool.OpenTag, tool.CloseTag)
			}
			continue
		}

		end := s + e + len(closeTag)
		block := data[s:end]
		h := sha256.Sum256(block)
		sha := hex.EncodeToString(h[:])

		if toStdout {
			fmt.Println(string(block))
		} else {
			filePath := filepath.Join(outDir, tool.Name+".md")
			if err := os.WriteFile(filePath, block, 0644); err != nil {
				p("[!] %s: Failed to write %s: %v", tool.Name, filePath, err)
				continue
			}

			p("[✓] %s (Tag: %s)", tool.Name, tool.OpenTag)
			p("    Offset: 0x%x, Size: %d bytes", s, len(block))
			p("    SHA256: %s", sha)
			p("    Saved to: %s", filePath)

			if len(origData) > 0 {
				origS := bytes.Index(origData, openTag)
				if origS >= 0 {
					origE := bytes.Index(origData[origS:], closeTag)
					if origE >= 0 {
						origBlock := origData[origS : origS+origE+len(closeTag)]
						if !bytes.Equal(origBlock, block) {
							origFilePath := filepath.Join(outDir, tool.Name+".original.md")
							_ = os.WriteFile(origFilePath, origBlock, 0644)
							origH := sha256.Sum256(origBlock)
							p("    Backup upstream extracted: %s (Size: %d bytes, SHA256: %s)", origFilePath, len(origBlock), hex.EncodeToString(origH[:]))
						}
					}
				}
			}
			p("")
		}
		extractedCount++
	}

	if !toStdout {
		p("Exported %d prompt(s) successfully.", extractedCount)
		p("════════════════════════════════════════════════════════════")
	}

	return out.String(), nil
}

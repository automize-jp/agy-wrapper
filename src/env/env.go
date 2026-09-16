package env

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func IsDirInPath(targetDir string) bool {
	cleanTarget, err := filepath.Abs(targetDir)
	if err != nil {
		cleanTarget = targetDir
	}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		absDir, err := filepath.Abs(dir)
		if err != nil {
			absDir = dir
		}
		if runtime.GOOS == "windows" {
			if strings.EqualFold(absDir, cleanTarget) {
				return true
			}
		} else {
			if absDir == cleanTarget {
				return true
			}
		}
	}

	if runtime.GOOS == "windows" {
		out, err := exec.Command("reg", "query", "HKCU\\Environment", "/v", "Path").Output()
		if err == nil {
			str := string(out)
			var pathVal string
			if idx := strings.Index(str, "REG_SZ"); idx != -1 {
				pathVal = strings.TrimSpace(str[idx+len("REG_SZ"):])
			} else if idx := strings.Index(str, "REG_EXPAND_SZ"); idx != -1 {
				pathVal = strings.TrimSpace(str[idx+len("REG_EXPAND_SZ"):])
			}
			if pathVal != "" {
				for _, part := range strings.Split(pathVal, ";") {
					part = strings.TrimSpace(part)
					if strings.EqualFold(filepath.Clean(part), cleanTarget) {
						return true
					}
				}
			}
		}
	}

	return false
}

func AddToUserPath(targetDir string) error {
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		absDir = targetDir
	}

	if runtime.GOOS == "windows" {
		psCmd := fmt.Sprintf(`$p = [Environment]::GetEnvironmentVariable("Path", "User"); $parts = $p -split ';' | Where-Object { $_ -ne '' -and $_ -ne '%s' }; [Environment]::SetEnvironmentVariable("Path", "%s;" + ($parts -join ';'), "User")`, absDir, absDir)
		cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
		return cmd.Run()
	}

	home, _ := os.UserHomeDir()
	shellConfig := filepath.Join(home, ".bashrc")
	if _, err := os.Stat(filepath.Join(home, ".zshrc")); err == nil {
		shellConfig = filepath.Join(home, ".zshrc")
	}

	f, err := os.OpenFile(shellConfig, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("\nexport PATH=\"%s:$PATH\"\n", absDir)
	_, err = f.WriteString(line)
	return err
}

// WrapperMarker identifies wrapper binaries
const WrapperMarker = "AGY_WRAPPER_SIGNATURE_MARKER_V1"

// IsCandidateWrapper returns true if candidate is another wrapper or non-upstream binary
func IsCandidateWrapper(candidatePath string) bool {
	fi, err := os.Stat(candidatePath)
	if err != nil || fi.IsDir() {
		return true
	}
	// Upstream Antigravity CLI binary is ~196 MB.
	// Any wrapper executable is small (typically 3-5 MB).
	// We safely reject any binary under 20 MB as a potential wrapper.
	if fi.Size() < 20*1024*1024 {
		return true
	}
	return false
}

func FindRealBinary(selfExe, selfDir string) (string, error) {
	dirs := filepath.SplitList(os.Getenv("PATH"))

	var extensions []string
	if runtime.GOOS == "windows" {
		pathext := os.Getenv("PATHEXT")
		if pathext == "" {
			pathext = ".COM;.EXE;.BAT;.CMD"
		}
		for _, ext := range strings.Split(pathext, ";") {
			if ext != "" {
				extensions = append(extensions, strings.ToLower(ext))
			}
		}
	} else {
		extensions = []string{""}
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		absDir, err := filepath.Abs(dir)
		if err != nil {
			absDir = dir
		}

		if runtime.GOOS == "windows" {
			if strings.EqualFold(absDir, selfDir) {
				continue
			}
		} else {
			if absDir == selfDir {
				continue
			}
		}

		for _, ext := range extensions {
			candidate := filepath.Join(absDir, "agy"+ext)
			evalCandidate, err := filepath.EvalSymlinks(candidate)
			if err != nil {
				continue
			}

			if runtime.GOOS == "windows" {
				if strings.EqualFold(evalCandidate, selfExe) {
					continue
				}
			} else {
				if evalCandidate == selfExe {
					continue
				}
			}

			fi, err := os.Stat(evalCandidate)
			if err != nil || fi.IsDir() {
				continue
			}

			if runtime.GOOS != "windows" && fi.Mode()&0111 == 0 {
				continue
			}

			if IsCandidateWrapper(evalCandidate) {
				continue
			}

			return evalCandidate, nil
		}
	}

	return "", fmt.Errorf("original 'agy' binary not found in PATH")
}

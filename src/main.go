package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"agy-wrapper/src/env"
	"agy-wrapper/src/meta"
	"agy-wrapper/src/patcher"
	"agy-wrapper/src/prompt"
)

func main() {
	selfExe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	selfExe, _ = filepath.EvalSymlinks(selfExe)
	selfDir := filepath.Dir(selfExe)

	// 1. Locate original binary in PATH
	realBin, err := env.FindRealBinary(selfExe, selfDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 2. Handle --extract-prompt flag
	for i, arg := range os.Args[1:] {
		if arg == "--extract-prompt" || strings.HasPrefix(arg, "--extract-prompt=") {
			var outDir string
			var filterTool string
			toStdout := false

			if strings.HasPrefix(arg, "--extract-prompt=") {
				outDir = strings.TrimPrefix(arg, "--extract-prompt=")
			}

			for idx, a := range os.Args[1:] {
				if a == "--stdout" {
					toStdout = true
				} else if strings.HasPrefix(a, "--tool=") {
					filterTool = strings.TrimPrefix(a, "--tool=")
				} else if a == "--tool" && idx+2 < len(os.Args) {
					filterTool = os.Args[1+idx+1]
				}
			}

			if outDir == "" {
				for _, nextArg := range os.Args[1+i+1:] {
					if !strings.HasPrefix(nextArg, "-") && nextArg != filterTool {
						outDir = nextArg
						break
					}
				}
			}

			logMsg, err := prompt.ExtractPrompts(realBin, outDir, toStdout, filterTool)
			if !toStdout {
				fmt.Print(logMsg)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// 3. Ensure selfDir is in PATH on first run
	if !env.IsDirInPath(selfDir) {
		fmt.Println("===================================================================")
		fmt.Println(" [AGY Wrapper] First run detected (directory not in PATH)!")
		fmt.Printf(" Automatically adding '%s' to PATH...\n", selfDir)

		if err := env.AddToUserPath(selfDir); err != nil {
			fmt.Fprintf(os.Stderr, " Error adding to PATH: %v\n", err)
		} else {
			fmt.Println(" ✓ Successfully added to PATH!")
		}

		fmt.Printf(" Target binary: %s\n", realBin)
		fmt.Println(" Running initial dry-run verification...")
		fmt.Println("-------------------------------------------------------------------")
		logMsg, err := patcher.PatchBinary(realBin, true)
		fmt.Print(logMsg)
		fmt.Println("===================================================================")
		if err != nil {
			fmt.Fprintf(os.Stderr, " Dry-run failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(" [AGY Wrapper] Setup completed successfully!")
		fmt.Println(" Please restart your terminal to apply the updated PATH and use 'agy'.")
		os.Exit(0)
	}

	// 4. Handle --dry-run flag using SHA256 fast-check
	for _, arg := range os.Args[1:] {
		if arg == "--dry-run" {
			metaFile := meta.GetMetaFile()
			var m meta.PatchMeta
			metaReadOk := false
			if data, err := os.ReadFile(metaFile); err == nil {
				if err := json.Unmarshal(data, &m); err == nil {
					metaReadOk = true
				}
			}
			curSHA, _ := meta.FileSHA256(realBin)
			targetPrompts := meta.GetTargetPromptsSHA256([]string{"grill-me"})
			targetPatches := meta.GetTargetStringPatchesSHA256()
			promptsOk := metaReadOk && meta.PromptsMatch(m.Prompts, targetPrompts) && meta.PromptsMatch(m.Patches, targetPatches)

			if metaReadOk && m.SHA256 != "" && curSHA == m.SHA256 {
				if promptsOk {
					fmt.Println("════════════════════════════════════════════════════════════")
					fmt.Printf("Binary: %s\n", realBin)
					fmt.Printf("SHA256: %s\n", curSHA)
					fmt.Println("Prompts: target fixed prompts match record ✓")
					fmt.Println("Status: SHA256 and prompt hash match patched record ✓ (no patch needed)")
					os.Exit(0)
				}
				fmt.Println("════════════════════════════════════════════════════════════")
				fmt.Printf("Binary: %s\n", realBin)
				fmt.Printf("SHA256: %s (patched with previous prompt version)\n", curSHA)
				fmt.Println("Status: Prompt .fixed changed (old patch detected). Simulating dry-run re-patch...")
				logMsg, err := patcher.PatchBinary(realBin, true)
				fmt.Print(logMsg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
					os.Exit(1)
				}
				os.Exit(0)
			}

			fmt.Println("════════════════════════════════════════════════════════════")
			fmt.Printf("Binary: %s\n", realBin)
			fmt.Printf("SHA256: %s (unpatched or upstream update detected)\n", curSHA)
			fmt.Println("Simulating dry-run patch...")
			logMsg, err := patcher.PatchBinary(realBin, true)
			fmt.Print(logMsg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// 5. Normal mode: ensure binary is patched silently
	meta.EnsurePatchedSilently(realBin)

	// 6. Forward arguments, stdio streams, and exit code to target binary
	cmd := exec.Command(realBin, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}

[🇺🇸 English](README.md) | [🇻🇳 Tiếng Việt](i18n/vi/README.md) | [🇰🇷 한국어](i18n/kr/README.md) | [🇨🇳 简体中文](i18n/cn/README.md) | [🇯🇵 日本語](i18n/ja/README.md)

# agy-wrapper

Transparent wrapper and automated binary patcher for the Google Antigravity CLI (`agy`).

## Overview

`agy-wrapper` automatically locates the upstream `agy` binary in your `PATH`, patches custom prompt definitions and in-place string replacements (supporting PE, ELF, and Mach-O executables), and caches metadata so that daily runs execute with zero latency overhead (~0ms).

## Features

- **Automated Patching**: Patches prompts and in-place binary strings directly in the original binary without corrupting ELF/PE/Mach-O layout.
- **Telemetry / Logging Opt-Out**: Replaces `https://play.googleapis.com/log` with `http://0.0.0.0/logging-disabled` in-place to disable Google Play logging.
- **Fail-Safe Verification**: Aborts cleanly if the upstream binary prompt has changed or checksum does not match.
- **Prompt Extraction**: Export embedded prompts directly from `agy` (`--extract-prompt`).
- **Fast Path Cache**: Stores SHA256, mod-time, and patch hash metadata in `~/.gemini/.latest-patched-version` for instant subsequent executions.
- **Transparent Proxy**: Preserves all CLI arguments, stdio streams (`stdin`, `stdout`, `stderr`), and process exit codes.
- **Automatic Setup**: Automatically registers its directory to user `PATH` on first run.

---

## Active Prompt Patches

`agy-wrapper` currently optimizes and patches the built-in `/grill-me` command (the interactive design & requirements interview agent):

- **Upstream Default (`grill-me.md`)**:
  - Asks questions strictly one at a time, requiring excessive interactive back-and-forth turns.
  - Hardcoded to English-only user communication.
  - Lacks trade-off explanations and actionable recommendations per question.

- **Patched Behavior (`grill-me.fixed.md`)**:
  - **Question Batching**: Groups independent questions together ordered by logical dependency to speed up alignment.
  - **Context & Recommendations**: Explains why each decision matters and proactively proposes recommended solutions with rationale.
  - **Conversational Language Matching**: Conducts all user-facing interviews in the language currently used in the conversation (Vietnamese, English, Japanese, etc.) and clarifies foreign technical terms when helpful.

> [!TIP]
> **Extensible to Other Prompts**: This binary patching and memory relocation approach is not limited to `/grill-me`. It can be readily applied to many other system prompts and slash commands within the Antigravity CLI (such as `/plan`, `/learn`, `/teamwork-preview`, `/browser`, etc.). You can inspect and extract any embedded prompt using `agy --extract-prompt`, customize it, and configure additional tools in the wrapper engine.

---

## Active String Patches

In addition to prompt patches, `agy-wrapper` applies equal-length in-place binary string replacements:

- **`disable-play-logging`**: Replaces all occurrences of `https://play.googleapis.com/log` with `http://0.0.0.0/logging-disabled` (31 bytes each) to disable telemetry/logging uploads. If the target string is already patched or absent in a build, it is skipped safely.

---

## Project Structure

```
├── i18n/                 # Multi-language documentation (vi, kr, cn, ja)
├── src/                  # Modular Go source code
│   ├── env/              # Environment PATH management & binary discovery
│   │   └── env.go
│   ├── meta/             # Metadata cache & verification
│   │   └── meta.go
│   ├── parser/           # Binary layout definitions & parsers (PE, ELF, Mach-O)
│   │   ├── binary.go
│   │   ├── pe.go
│   │   ├── elf.go
│   │   └── macho.go
│   ├── patcher/          # Binary patching engine & memory relocation
│   │   └── patcher.go
│   ├── prompt/           # Embedded prompts & extraction engine
│   │   ├── prompts/      # Embedded prompt markdown definitions
│   │   └── prompt.go
│   └── main.go           # CLI entry point, argument routing & proxy
├── go.mod                # Go module specification
├── LICENSE               # MIT License
└── README.md             # English documentation
```

---

## Pre-built Releases

Pre-compiled binary ZIP archives are automatically built and published on GitHub Releases when a version tag (e.g. `v1.0.0`) is pushed:

| Platform | Archive Name | Binary Inside |
| :--- | :--- | :--- |
| **Windows x86_64** | `agy-wrapper-v{VERSION}-windows-amd64.zip` | `agy.exe` |
| **Linux x86_64** | `agy-wrapper-v{VERSION}-linux-amd64.zip` | `agy` |
| **Linux ARM64** | `agy-wrapper-v{VERSION}-linux-arm64.zip` | `agy` |
| **macOS Intel** | `agy-wrapper-v{VERSION}-darwin-amd64.zip` | `agy` |
| **macOS Apple Silicon** | `agy-wrapper-v{VERSION}-darwin-arm64.zip` | `agy` |
| **macOS Universal** | `agy-wrapper-v{VERSION}-darwin-universal.zip` | `agy` |

*Each ZIP archive contains the executable binary, `LICENSE`, `README.md`, and the `i18n/` documentation directory.*

---

## Build Instructions

All build outputs are directed to the `dist/` directory (excluded from Git).

Requires **Go 1.21+**.

### Windows

**On Windows (PowerShell):**
```powershell
go build -ldflags="-s -w" -o dist/agy.exe ./src
```

**Cross-compile from macOS/Linux:**
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy.exe ./src
```

---

### Linux

**On Linux or cross-compile (x86_64):**
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy ./src
```

**ARM64 (Raspberry Pi, AWS Graviton, etc.):**
```bash
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy ./src
```

---

### macOS

**Apple Silicon (M1/M2/M3/M4):**
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy ./src
```

**Intel Mac (x86_64):**
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy ./src
```

**Universal Binary (Apple Silicon + Intel):**
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy-arm64 ./src
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy-amd64 ./src
lipo -create -output dist/agy dist/agy-arm64 dist/agy-amd64
rm dist/agy-arm64 dist/agy-amd64
```

---

## Safety & Prompt Mismatch Behavior

When `agy-wrapper` runs, it scans the upstream binary for the expected prompt:
1. **Exact Checksum Match**: The patch is applied only if the target prompt inside the binary exactly matches the expected original prompt.
2. **If Upstream Prompt Changed**:
   - The patcher aborts with `Checksum mismatch — upstream prompt has changed` or `Prompt not found in binary`.
   - **The original binary is never modified or corrupted** (all checks run in memory before applying changes).
   - An automatic backup (`<binary>.original`) is preserved before any disk write.
3. **Updating Prompts**:
   - Prompts are loaded from disk (`./prompts/<platform>/` or `./src/prompts/<platform>/`) first, falling back to embedded binaries.
   - If Google releases a new version of `agy`, you can simply drop the new `grill-me.md` and `grill-me.fixed.md` into `./prompts/` or recompile with updated assets.

---

## Usage

1. Copy the compiled executable from `dist/` into a folder of your choice.
2. Run once to trigger initial setup and dry-run verification:
   ```bash
   agy
   ```
3. Restart your terminal to pick up updated `PATH`, then use `agy` as normal:
   ```bash
   agy --help
   ```

### Dry-Run Verification

Check binary status and simulate patching without writing to disk:
```bash
agy --dry-run
```

### Extract Prompts from `agy`

Export embedded prompts directly from the upstream `agy` binary:
```bash
# Export all prompts into prompts/<platform>/ by default
agy --extract-prompt

# Export to a custom output directory
agy --extract-prompt ./my_prompts

# Export a specific tool prompt (e.g. grill-me, plan, learn)
agy --extract-prompt --tool=grill-me

# Print prompt content directly to stdout
agy --extract-prompt --tool=grill-me --stdout
```

When exporting, if an `.original` backup binary exists alongside the target executable, `agy-wrapper` will automatically detect it and export pristine upstream versions (e.g. `<tool>.original.md`).

---

## License

This project is licensed under the [MIT License](LICENSE).

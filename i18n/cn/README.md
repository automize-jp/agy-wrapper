[🇺🇸 English](../../README.md) | [🇻🇳 Tiếng Việt](../vi/README.md) | [🇰🇷 한국어](../kr/README.md) | [🇨🇳 简体中文](README.md) | [🇯🇵 日本語](../ja/README.md)

# agy-wrapper

适用于 Google Antigravity CLI (`agy`) 的透明封装器与自动化二进制补丁工具。

## 概述

`agy-wrapper` 能够自动定位环境变量 `PATH` 中的上游 `agy` 二进制文件，直接对可执行文件内的提示词（Prompt）与等长字符串进行就地补丁（支持 PE、ELF 和 Mach-O 格式），并通过缓存元数据实现后续运行零延迟（~0ms）透明转发。

## 功能特性

- **自动化二进制补丁**：直接对二进制内部的 Prompt 与字符串进行替换，不破坏 PE/ELF/Mach-O 的内存布局与段结构。
- **遥测 / 日志禁用**：将 `https://play.googleapis.com/log` 就地替换为 `http://0.0.0.0/logging-disabled`，禁用 Google Play 日志上报。
- **安全失败保障（Fail-Safe）**：当上游二进制被官方更新或校验和不匹配时，自动中止操作，杜绝文件损坏。
- **Prompt 导出提取**：支持直接从二进制文件中提取内置系统提示词（`--extract-prompt`）。
- **极速缓存（Fast Path Cache）**：在 `~/.gemini/.latest-patched-version` 中缓存 SHA256、文件修改时间及补丁哈希，启动无感。
- **透明代理**：完整保留所有命令行参数、标准输入输出流（`stdin`、`stdout`、`stderr`）以及进程退出码。
- **自动环境配置**：首次运行时自动将当前目录加入用户 `PATH` 环境变量。

---

## 正在生效的 Prompt 补丁 (Active Patches)

`agy-wrapper` 目前默认对内置的 `/grill-me` 命令（任务启动前澄清需求与架构设计的访谈代理）进行深度优化与补丁：

* **官方默认行为 (`grill-me.md`)**：
  * 严格限制一次只能提一个问题，产生极其冗长的前后轮次等待。
  * 强制仅使用英文进行交互。
  * 缺乏决策背景与权衡说明，不主动提供推荐方案。

* **补丁后优化行为 (`grill-me.fixed.md`)**：
  * **问题分批合并 (Question Batching)**：按逻辑依赖关系将独立问题分组，一次性合并提出，大幅节省沟通时间。
  * **背景说明与推荐建议**：详述每个决策点的重要意义及权衡因素，并主动提供附带技术理由的推荐答案。
  * **对话语言自动匹配 (Conversational Language Matching)**：全自动跟随当前对话所使用的语种（中文、英文、越南语、日语等）进行提问；使用外来专业技术词汇时附带简要中文解释。

> [!TIP]
> **可轻松扩展至其他 Prompt**：该二进制补丁与内存重定位技术绝不仅限于 `/grill-me`，完全适用于 Antigravity CLI 的众多系统提示词和斜杠命令（如 `/plan`、`/learn`、`/teamwork-preview`、`/browser` 等）。您可以随时使用 `agy --extract-prompt` 提取感兴趣的内置提示词，按需定制并在 wrapper 引擎中扩展注册。

---

## 正在生效的字符串补丁 (Active String Patches)

除 Prompt 补丁外，`agy-wrapper` 还会对二进制中的等长字符串执行就地替换：

- **`disable-play-logging`**：将所有出现的 `https://play.googleapis.com/log` 替换为 `http://0.0.0.0/logging-disabled`（均为 31 字节），以禁用遥测/日志上传。若目标字符串已被补丁或当前构建中不存在，则会安全跳过。

---

## 项目结构

```
├── i18n/                 # 多语言文档 (vi, kr, cn, ja)
├── src/                  # 按功能拆分的 Go 模块化源码
│   ├── env/              # 环境变量 PATH 管理与二进制查找
│   │   └── env.go
│   ├── meta/             # 元数据缓存与版本校验
│   │   └── meta.go
│   ├── parser/           # 二进制布局与格式解析器 (PE, ELF, Mach-O)
│   │   ├── binary.go
│   │   ├── pe.go
│   │   ├── elf.go
│   │   └── macho.go
│   ├── patcher/          # 补丁应用引擎与内存重定位
│   │   └── patcher.go
│   ├── prompt/           # 内置 Prompt 管理与提取引擎
│   │   ├── prompts/      # 内置 Prompt Markdown 文件
│   │   └── prompt.go
│   └── main.go           # CLI 入口、参数路由与透明代理
├── go.mod                # Go 模块配置文件
├── LICENSE               # MIT 开源许可证
└── README.md             # 英文主文档
```

---

## 预构建发行包 (Releases)

当推送版本标签（例如 `v1.0.0`）时，GitHub Releases 将自动构建并发布预编译的二进制 ZIP 压缩包：

| 平台架构 | 压缩包名称 | 包含的可执行文件 |
| :--- | :--- | :--- |
| **Windows x86_64** | `agy-wrapper-v{VERSION}-windows-amd64.zip` | `agy.exe` |
| **Linux x86_64** | `agy-wrapper-v{VERSION}-linux-amd64.zip` | `agy` |
| **Linux ARM64** | `agy-wrapper-v{VERSION}-linux-arm64.zip` | `agy` |
| **macOS Intel** | `agy-wrapper-v{VERSION}-darwin-amd64.zip` | `agy` |
| **macOS Apple Silicon** | `agy-wrapper-v{VERSION}-darwin-arm64.zip` | `agy` |
| **macOS Universal** | `agy-wrapper-v{VERSION}-darwin-universal.zip` | `agy` |

*每个 ZIP 压缩包均包含对应的可执行二进制文件、`LICENSE`、`README.md` 以及完整的 `i18n/` 多语言文档目录。*

---

## 构建指南

所有构建产物默认输出至 `dist/` 目录（已在 Git 中忽略）。

编译要求：**Go 1.21+**。

### Windows

**在 Windows 本地编译（PowerShell）：**
```powershell
go build -ldflags="-s -w" -o dist/agy.exe ./src
```

**从 macOS/Linux 交叉编译为 Windows：**
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy.exe ./src
```

---

### Linux

**在 Linux 本地或交叉编译（x86_64）：**
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy ./src
```

**ARM64 架构（树莓派、AWS Graviton 等）：**
```bash
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy ./src
```

---

### macOS

**Apple Silicon 芯片（M1/M2/M3/M4）：**
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy ./src
```

**Intel 芯片（x86_64）：**
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy ./src
```

**通用二进制文件（Universal Binary，同时支持 Apple Silicon 与 Intel）：**
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy-arm64 ./src
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy-amd64 ./src
lipo -create -output dist/agy dist/agy-arm64 dist/agy-amd64
rm dist/agy-arm64 dist/agy-amd64
```

---

## 提示词不匹配与安全机制

当 `agy-wrapper` 运行时，会严格校验二进制内容：
1. **精确校验和匹配**：仅在目标二进制内的原始提示词与预期字节完全一致时才执行补丁。
2. **上游官方更新提示词时**：
   - 补丁程序会输出 `Checksum mismatch — upstream prompt has changed` 并立即终止。
   - **原始二进制绝对不会被破坏或篡改**（所有比对均在内存中完成）。
   - 在任何写盘操作前均会自动创建 `<binary>.original` 备份文件。
3. **更新提示词**：
   - 程序优先从磁盘目录（`prompts/<platform>/` 或 `src/prompts/<platform>/`）加载提示词，随后回退至嵌入资源。
   - 官方更新后，只需将新的提示词提取并放置在 `prompts/` 目录下即可无缝更新。

---

## 使用说明

1. 将编译生成在 `dist/` 中的可执行文件移动至目标目录。
2. 运行一次以自动配置环境变量 `PATH` 并完成自检：
   ```bash
   agy
   ```
3. 重启终端以使环境变量生效，之后即可像原生 `agy` 一样使用：
   ```bash
   agy --help
   ```

### 模拟运行（Dry-Run）

检查二进制状态并模拟补丁过程，不写入磁盘：
```bash
agy --dry-run
```

### 提取内置 Prompt

从目标 `agy` 二进制文件中导出内置系统提示词：
```bash
# 默认导出全部提示词至 prompts/<platform>/
agy --extract-prompt

# 导出至指定自定义目录
agy --extract-prompt ./my_prompts

# 仅导出指定工具的提示词（例如 grill-me, plan, learn）
agy --extract-prompt --tool=grill-me

# 直接打印至标准输出（stdout）
agy --extract-prompt --tool=grill-me --stdout
```

若目标路径存在备份文件（`.original`），程序会自动同步导出官方原版提示词（`<tool>.original.md`）以便对比。

---

## 许可证

本项目基于 [MIT License](../../LICENSE) 开源。

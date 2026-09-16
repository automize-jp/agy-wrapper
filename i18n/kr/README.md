[🇺🇸 English](../../README.md) | [🇻🇳 Tiếng Việt](../vi/README.md) | [🇰🇷 한국어](README.md) | [🇨🇳 简体中文](../cn/README.md) | [🇯🇵 日本語](../ja/README.md)

# agy-wrapper

Google Antigravity CLI (`agy`)를 위한 투명 래퍼 및 자동 바이너리 패처.

## 개요

`agy-wrapper`는 시스템 `PATH`에서 원본 `agy` 바이너리를 자동으로 감지하고, 커스텀 프롬프트 정의를 바이너리에 직접 패치(PE, ELF, Mach-O 지원)하며, 메타데이터를 캐싱하여 일상적인 실행 시 지연 시간 0ms(~0ms)로 투명하게 프록시 실행합니다.

## 주요 기능

- **자동 바이너리 패치**: PE/ELF/Mach-O의 메모리 레이아웃이나 세그먼트 손상 없이 바이너리 내부 프롬프트를 안전하게 패치합니다.
- **안전 검증 (Fail-Safe)**: 업스트림 바이너리의 프롬프트가 변경되었거나 체크섬이 일치하지 않을 경우 안전하게 작업을 중단합니다.
- **프롬프트 추출**: `agy` 바이너리에서 내장된 시스템 프롬프트를 직접 추출합니다 (`--extract-prompt`).
- **고속 캐시 (Fast Path Cache)**: SHA256 및 수정 시간 메타데이터를 `~/.gemini/.latest-patched-version`에 저장하여 즉시 실행합니다.
- **투명 프록시 (Transparent Proxy)**: 모든 CLI 인자, 표준 입출력 스트림(`stdin`, `stdout`, `stderr`) 및 종료 코드를 그대로 보존합니다.
- **자동 환경 설정**: 최초 실행 시 현재 디렉터리를 사용자 `PATH`에 자동으로 등록합니다.

---

## 현재 적용된 프롬프트 패치 (Active Patches)

`agy-wrapper`는 내장된 `/grill-me` 명령어(작업 전 요구사항 및 설계를 조율하는 인터뷰 에이전트)를 자동으로 최적화 및 패치합니다:

* **업스트림 기본 동작 (`grill-me.md`)**:
  * 한 번에 질문을 하나씩만 던지도록 강제되어 있어 수많은 대화 턴이 소모됩니다.
  * 영어로만 대화하도록 고정되어 있습니다.
  * 설계 트레이드오프 분석 및 권장 해결책 제시가 부족합니다.

* **패치 후 개선된 동작 (`grill-me.fixed.md`)**:
  * **질문 묶음 전송 (Question Batching)**: 독립적인 질문들을 의존성 순서에 따라 그룹으로 묶어 한 번에 질문함으로써 소통 시간을 크게 줄입니다.
  * **맥락 분석 및 해결책 제안**: 각 결정 사항이 왜 중요한지 설명하고, 기술적 근거와 함께 권장 해결책을 능동적으로 제안합니다.
  * **대화 언어 자동 일치 (Conversational Language Matching)**: 현재 대화에서 사용 중인 언어(한국어, 영어, 베트남어, 일본어 등)로 모든 질문을 자동으로 진행하며, 타 언어의 전문 기술 용어 사용 시 간결한 설명을 함께 제공합니다.

> [!TIP]
> **다양한 프롬프트로의 확장 가능성**: 이 바이너리 패치 및 메모리 재배치 기법은 `/grill-me`에만 국한되지 않으며, Antigravity CLI의 다양한 시스템 프롬프트 및 슬래시 명령어(`/plan`, `/learn`, `/teamwork-preview`, `/browser` 등)에도 즉시 응용될 수 있습니다. `agy --extract-prompt`로 원하는 프롬프트를 추출하여 커스텀한 후 래퍼 엔진에 도구를 추가 등록하면 됩니다.

---

## 프로젝트 구조

```
├── i18n/                 # 다국어 문서 (vi, kr, cn, ja)
├── src/                  # 기능별로 모듈화된 Go 소스 코드
│   ├── env/              # 환경 변수 PATH 관리 및 바이너리 탐색
│   │   └── env.go
│   ├── meta/             # 메타데이터 캐시 및 유효성 검사
│   │   └── meta.go
│   ├── parser/           # 바이너리 레이아웃 및 파서 (PE, ELF, Mach-O)
│   │   ├── binary.go
│   │   ├── pe.go
│   │   ├── elf.go
│   │   └── macho.go
│   ├── patcher/          # 바이너리 패치 엔진 및 메모리 재배치
│   │   └── patcher.go
│   ├── prompt/           # 임베디드 프롬프트 관리 및 추출 엔진
│   │   ├── prompts/      # 임베디드 프롬프트 마크다운 파일
│   │   └── prompt.go
│   └── main.go           # CLI 진입점, 인자 라우팅 및 프록시
├── go.mod                # Go 모듈 파일
├── LICENSE               # MIT 라이선스
└── README.md             # 영문 기본 문서
```

---

## 사전 빌드된 릴리스 (Releases)

버전 태그(예: `v1.0.0`)가 푸시되면 GitHub Releases에 사전 컴파일된 바이너리 ZIP 아카이브가 자동으로 생성 및 배포됩니다:

| 플랫폼 | ZIP 아카이브 파일명 | 내부 바이너리 |
| :--- | :--- | :--- |
| **Windows x86_64** | `agy-wrapper-v{VERSION}-windows-amd64.zip` | `agy.exe` |
| **Linux x86_64** | `agy-wrapper-v{VERSION}-linux-amd64.zip` | `agy` |
| **Linux ARM64** | `agy-wrapper-v{VERSION}-linux-arm64.zip` | `agy` |
| **macOS Intel** | `agy-wrapper-v{VERSION}-darwin-amd64.zip` | `agy` |
| **macOS Apple Silicon** | `agy-wrapper-v{VERSION}-darwin-arm64.zip` | `agy` |
| **macOS Universal** | `agy-wrapper-v{VERSION}-darwin-universal.zip` | `agy` |

*각 ZIP 아카이브에는 실행 파일, `LICENSE`, `README.md` 및 `i18n/` 다국어 문서 폴더가 포함됩니다.*

---

## 빌드 방법

모든 빌드 결과물은 `dist/` 디렉터리에 저장됩니다 (Git 제외 대상).

요구 사양: **Go 1.21+**.

### Windows

**Windows 네이티브 (PowerShell):**
```powershell
go build -ldflags="-s -w" -o dist/agy.exe ./src
```

**macOS/Linux에서 Windows용 크로스 컴파일:**
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy.exe ./src
```

---

### Linux

**Linux 네이티브 또는 크로스 컴파일 (x86_64):**
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy ./src
```

**ARM64 (라즈베리 파이, AWS Graviton 등):**
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

**유니버설 바이너리 (Apple Silicon + Intel 결합):**
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy-arm64 ./src
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy-amd64 ./src
lipo -create -output dist/agy dist/agy-arm64 dist/agy-amd64
rm dist/agy-arm64 dist/agy-amd64
```

---

## 프롬프트 불일치 시 안전 대책

`agy-wrapper`는 패치 적용 전 원본 바이너리를 철저하게 검사합니다:
1. **정확한 체크섬 일치**: 대상 바이너리 내부의 프롬프트가 예상한 원본과 바이트 단위까지 정확히 일치할 때만 패치를 적용합니다.
2. **업스트림 프롬프트가 변경된 경우**:
   - `Checksum mismatch — upstream prompt has changed` 또는 `Prompt not found in binary` 오류와 함께 안전하게 중단됩니다.
   - **원본 바이너리는 손상되거나 수정되지 않습니다** (모든 사전 검증은 메모리 상에서만 수행됩니다).
   - 실제 디스크 쓰기 전 자동 백업(`*.original`)이 항상 생성됩니다.
3. **프롬프트 업데이트 방법**:
   - 프롬프트는 임베디드 리소스보다 로컬 디스크(`prompts/<platform>/` 또는 `src/prompts/<platform>/`) 경로를 먼저 참조합니다.
   - 새 버전이 출시되면 업데이트된 프롬프트를 디스크 경로에 추가하거나 바이너리를 다시 빌드하여 즉시 적용할 수 있습니다.

---

## 사용법

1. `dist/` 디렉터리에 생성된 실행 파일을 원하는 위치로 이동합니다.
2. 한 번 실행하여 PATH 자동 등록 및 드라이 런 검증을 진행합니다:
   ```bash
   agy
   ```
3. 터미널을 다시 열어 변경된 PATH를 적용한 뒤 일반 `agy` 명령어처럼 사용합니다:
   ```bash
   agy --help
   ```

### 드라이 런 (Dry-Run) 검증

디스크에 쓰지 않고 바이너리 상태 및 패치 시뮬레이션을 수행합니다:
```bash
agy --dry-run
```

### `agy`에서 프롬프트 추출

바이너리에 내장된 프롬프트를 로컬 파일로 내보냅니다:
```bash
# 기본 경로(prompts/<platform>/)로 모든 프롬프트 추출
agy --extract-prompt

# 특정 출력 디렉터리로 추출
agy --extract-prompt ./my_prompts

# 특정 도구 프롬프트만 추출 (예: grill-me, plan, learn)
agy --extract-prompt --tool=grill-me

# 표준 출력(stdout)으로 직접 출력
agy --extract-prompt --tool=grill-me --stdout
```

백업 바이너리(`.original`)가 함께 존재하는 경우 원본 업스트림 프롬프트(`<tool>.original.md`)도 자동으로 분리 추출됩니다.

---

## 라이선스

이 프로젝트는 [MIT 라이선스](../../LICENSE)를 따릅니다.

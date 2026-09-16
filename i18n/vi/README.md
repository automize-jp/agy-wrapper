[🇺🇸 English](../../README.md) | [🇻🇳 Tiếng Việt](README.md) | [🇰🇷 한국어](../kr/README.md) | [🇨🇳 简体中文](../cn/README.md) | [🇯🇵 日本語](../ja/README.md)

# agy-wrapper

Wrapper thông minh và công cụ tự động vá binary cho Google Antigravity CLI (`agy`).

## Tổng quan

`agy-wrapper` tự động định vị binary `agy` gốc trong biến môi trường `PATH`, vá các prompt hệ thống theo định dạng tùy chỉnh (hỗ trợ các định dạng PE, ELF, và Mach-O), đồng thời lưu bộ đệm metadata để các lần chạy tiếp theo đạt độ trễ xấp xỉ 0ms (~0ms).

## Tính năng

- **Tự động vá binary**: Vá trực tiếp nội dung prompt bên trong binary gốc mà không làm sai lệch cấu trúc segment hay offset của PE/ELF/Mach-O.
- **Cơ chế an toàn (Fail-Safe)**: Tự động hủy thao tác nếu prompt trong binary gốc đã thay đổi hoặc không khớp mã hash SHA256.
- **Trích xuất Prompt**: Xuất trực tiếp các prompt hệ thống từ binary `agy` (`--extract-prompt`).
- **Bộ nhớ đệm siêu tốc (Fast Path Cache)**: Lưu trữ metadata (SHA256, kích thước, thời gian sửa đổi) tại `~/.gemini/.latest-patched-version` để khởi động tức thì.
- **Chuyển tiếp trong suốt (Transparent Proxy)**: Giữ nguyên toàn bộ tham số dòng lệnh, các luồng I/O (`stdin`, `stdout`, `stderr`), và mã thoát (exit code).
- **Tự động cấu hình**: Tự động thêm thư mục wrapper vào `PATH` của người dùng ngay trong lần chạy đầu tiên.

---

## Các Prompt đang được Patch (Active Patches)

Hiện tại, `agy-wrapper` tự động phát hiện và tối ưu lệnh `/grill-me` (chế độ phỏng vấn làm rõ yêu cầu và thiết kế kiến trúc trước khi triển khai):

* **Bản gốc từ Google (`grill-me.md`)**:
  * Bắt buộc hỏi từng câu một đơn lẻ (`Ask the questions one at a time`), tốn rất nhiều lượt hội thoại chờ đợi.
  * Cố định giao tiếp bằng tiếng Anh.
  * Thiếu phần phân tích đánh đổi (trade-offs) và không chủ động đề xuất giải pháp.

* **Bản đã được vá (`grill-me.fixed.md`)**:
  * **Hỏi theo nhóm (Question Batching)**: Gom các câu hỏi độc lập để hỏi cùng lúc theo nhóm, sắp xếp theo thứ tự phụ thuộc logic giúp rút ngắn tối đa thời gian trao đổi.
  * **Phân tích bối cảnh & Đề xuất phương án**: Với mỗi câu hỏi, nêu rõ lý do tại sao nó quan trọng, các đánh đổi liên quan và chủ động đề xuất giải pháp khuyến nghị kèm lý do kỹ thuật.
  * **Thích ứng theo ngôn ngữ hội thoại (Conversational Language Matching)**: Tự động đặt câu hỏi theo đúng ngôn ngữ đang trao đổi trong cuộc hội thoại (tiếng Việt, tiếng Anh, tiếng Nhật, tiếng Hàn,...); khi dùng thuật ngữ từ ngôn ngữ khác sẽ có phần giải thích nghĩa ngắn gọn đi kèm.

> [!TIP]
> **Khả năng mở rộng cho nhiều prompt khác**: Phương pháp vá nhị phân và dời bảng con trỏ chuỗi này không chỉ giới hạn ở `/grill-me`, mà còn có thể áp dụng hiệu quả cho rất nhiều prompt và slash command khác của Antigravity CLI (như `/plan`, `/learn`, `/teamwork-preview`, `/browser`,...). Bạn có thể trích xuất bất kỳ prompt nào bằng lệnh `agy --extract-prompt`, tùy chỉnh nội dung và đăng ký thêm tool trong engine của wrapper.

---

## Cấu trúc dự án

```
├── i18n/                 # Tài liệu đa ngôn ngữ (vi, kr, cn, ja)
├── src/                  # Mã nguồn Go phân theo chức năng
│   ├── env/              # Quản lý biến môi trường PATH & dò tìm binary
│   │   └── env.go
│   ├── meta/             # Bộ nhớ đệm metadata & xác thực phiên bản
│   │   └── meta.go
│   ├── parser/           # Cấu trúc nhị phân & bộ phân giải PE, ELF, Mach-O
│   │   ├── binary.go
│   │   ├── pe.go
│   │   ├── elf.go
│   │   └── macho.go
│   ├── patcher/          # Engine vá binary & tái cấp phát bộ nhớ
│   │   └── patcher.go
│   ├── prompt/           # Quản lý prompt nhúng & engine trích xuất
│   │   ├── prompts/      # File markdown prompt nhúng
│   │   └── prompt.go
│   └── main.go           # Điểm khởi đầu CLI, điều hướng cờ lệnh & proxy
├── go.mod                # Định nghĩa module Go
├── LICENSE               # Giấy phép MIT
└── README.md             # Tài liệu tiếng Anh
```

---

## Bản phát hành đóng gói sẵn (Releases)

Các gói file ZIP biên dịch sẵn được tự động đóng gói và xuất bản trên GitHub Releases khi có tag phiên bản (ví dụ: `v1.0.0`):

| Nền tảng | Tên file ZIP | File nhị phân bên trong |
| :--- | :--- | :--- |
| **Windows x86_64** | `agy-wrapper-v{VERSION}-windows-amd64.zip` | `agy.exe` |
| **Linux x86_64** | `agy-wrapper-v{VERSION}-linux-amd64.zip` | `agy` |
| **Linux ARM64** | `agy-wrapper-v{VERSION}-linux-arm64.zip` | `agy` |
| **macOS Intel** | `agy-wrapper-v{VERSION}-darwin-amd64.zip` | `agy` |
| **macOS Apple Silicon** | `agy-wrapper-v{VERSION}-darwin-arm64.zip` | `agy` |
| **macOS Universal** | `agy-wrapper-v{VERSION}-darwin-universal.zip` | `agy` |

*Mỗi file ZIP đều chứa file thực thi, `LICENSE`, `README.md` và thư mục tài liệu `i18n/`.*

---

## Hướng dẫn Build

Toàn bộ file binary khi build sẽ được xuất vào thư mục `dist/` (đã được loại trừ khỏi Git).

Yêu cầu **Go 1.21+**.

### Windows

**Trên Windows (PowerShell):**
```powershell
go build -ldflags="-s -w" -o dist/agy.exe ./src
```

**Cross-compile từ macOS/Linux sang Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy.exe ./src
```

---

### Linux

**Trên Linux hoặc cross-compile (x86_64):**
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy ./src
```

**ARM64 (Raspberry Pi, AWS Graviton,...):**
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

**Universal Binary (chạy cả Apple Silicon + Intel):**
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy-arm64 ./src
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy-amd64 ./src
lipo -create -output dist/agy dist/agy-arm64 dist/agy-amd64
rm dist/agy-arm64 dist/agy-amd64
```

---

## Cơ chế an toàn khi Prompt không khớp

Khi `agy-wrapper` chạy, hệ thống sẽ kiểm tra binary gốc:
1. **Khớp mã hash chính xác (Exact Checksum)**: Thao tác vá chỉ được áp dụng khi chuỗi prompt gốc trong binary khớp chính xác từng byte.
2. **Nếu Prompt gốc của Google thay đổi**:
   * Quá trình vá sẽ dừng lại kèm thông báo: `Checksum mismatch — upstream prompt has changed` hoặc `Prompt not found in binary`.
   * **Binary gốc không bao giờ bị chỉnh sửa hay làm hỏng** (toàn bộ kiểm tra diễn ra trên RAM trước khi ghi đĩa).
   * File sao lưu tự động (`<binary>.original`) luôn được tạo trước khi ghi đè.
3. **Cập nhật Prompt mới**:
   * Prompt được ưu tiên đọc từ ổ đĩa (`./prompts/<platform>/` hoặc `./src/prompts/<platform>/`) trước khi dùng prompt nhúng mặc định.
   * Khi Google ra phiên bản `agy` mới, bạn chỉ cần cập nhật file `grill-me.md` và `grill-me.fixed.md` đặt vào thư mục `./prompts/` hoặc build lại.

---

## Hướng dẫn sử dụng

1. Copy file thực thi từ thư mục `dist/` vào vị trí bạn muốn lưu trữ.
2. Chạy thử một lần để kích hoạt tự động cấu hình PATH:
   ```bash
   agy
   ```
3. Mở cửa sổ Terminal mới để nhận PATH cập nhật, sau đó sử dụng `agy` như bình thường:
   ```bash
   agy --help
   ```

### Chạy kiểm tra Dry-Run

Kiểm tra trạng thái binary và mô phỏng vá mà không ghi thay đổi lên đĩa:
```bash
agy --dry-run
```

### Trích xuất Prompt từ `agy`

Xuất trực tiếp prompt hệ thống từ binary gốc của `agy`:
```bash
# Xuất toàn bộ prompt mặc định vào thư mục prompts/<platform>/
agy --extract-prompt

# Xuất vào thư mục chỉ định tùy ý
agy --extract-prompt ./my_prompts

# Chỉ xuất prompt của một tool cụ thể (grill-me, plan, learn)
agy --extract-prompt --tool=grill-me

# In trực tiếp nội dung prompt ra màn hình console / stdout
agy --extract-prompt --tool=grill-me --stdout
```

Khi xuất, nếu phát hiện file backup `.original` nằm cạnh binary mục tiêu, `agy-wrapper` sẽ tự động trích xuất thêm bản gốc sạch của Google (ví dụ `<tool>.original.md`).

---

## Giấy phép

Dự án được phát hành theo giấy phép [MIT License](../../LICENSE).

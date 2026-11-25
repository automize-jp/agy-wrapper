[🇺🇸 English](../../README.md) | [🇻🇳 Tiếng Việt](../vi/README.md) | [🇰🇷 한국어](../kr/README.md) | [🇨🇳 简体中文](../cn/README.md) | [🇯🇵 日本語](README.md)

# agy-wrapper

Google Antigravity CLI (`agy`) のための透過的ラッパーおよび自動バイナリパッチャー。

## 概要

`agy-wrapper` はシステム `PATH` 内の公式 `agy` バイナリを自動検出し、バイナリ内のシステムプロンプト定義やインプレース文字列置換を直接パッチ（PE、ELF、Mach-O 形式対応）します。メタデータのキャッシュにより、通常実行時は遅延約 0ms（~0ms）で透過的にプロキシ実行されます。

## 主な機能

- **自動バイナリパッチ**：PE/ELF/Mach-O のセグメントやレイアウト構造を壊すことなく、バイナリ内のプロンプトや文字列をインプレースで安全にパッチします。
- **テレメトリ / ロギング無効化**：`https://play.googleapis.com/log` を `http://0.0.0.0/logging-disabled` にインプレース置換し、Google Play ロギングを無効化します。
- **フェイルセーフ機構（Fail-Safe）**：上流のバイナリが更新された場合やチェックサムが一致しない場合、処理を安全に中断します。
- **プロンプト抽出機能**：`agy` バイナリから内蔵システムプロンプトを直接ファイルへエクスポート（`--extract-prompt`）。
- **高速メタデータキャッシュ**：SHA256、更新日時、およびパッチハッシュを `~/.gemini/.latest-patched-version` にキャッシュし、瞬時に起動します。
- **完全透過プロキシ**：コマンドライン引数、標準入出力ストリーム（`stdin`、`stdout`、`stderr`）、終了コードを完全に保持します。
- **自動環境変数設定**：初回実行時に現在のディレクトリをユーザー環境変数 `PATH` へ自動登録します。

---

## 現在適用されているプロンプトパッチ (Active Patches)

`agy-wrapper` は現在、内蔵の `/grill-me` コマンド（実装前の要件定義および設計すり合わせを行うインタビューエージェント）を自動的に最適化・パッチします：

* **公式デフォルトの挙動 (`grill-me.md`)**：
  * 厳格に1問ずつ質問するため、膨大な会話往復と待ち時間が発生します。
  * 英語での会話に固定されています。
  * 各設問に対するトレードオフ分析や具体的な推奨案の提示が不足しています。

* **パッチ適用後の最適化された挙動 (`grill-me.fixed.md`)**：
  * **質問のバッチ送信 (Question Batching)**：論理的依存関係に基づいて独立した質問をグループ化し、一度にまとめて提示することでやり取りの時間を劇的に短縮します。
  * **背景説明と推奨案の能動的提示**：なぜ各設問が重要なのか理由とトレードオフを明記し、技術的根拠を添えた推奨回答を提示します。
  * **会話言語への自動追従 (Conversational Language Matching)**：現在会話で使用されている言語（日本語、英語、ベトナム語など）で自動的に質問を行い、他言語の専門用語を使用する際は簡潔な補足解説を添えます。

> [!TIP]
> **他のプロンプトへの容易な拡張性**：このバイナリパッチおよびメモリ再配置技術は `/grill-me` だけにとどまらず、Antigravity CLI の他の多くのシステムプロンプトやスラッシュコマンド（`/plan`、`/learn`、`/teamwork-preview`、`/browser` など）にも容易に応用可能です。`agy --extract-prompt` でプロンプトを抽出してカスタマイズし、wrapper エンジンにツールを追加登録することで柔軟に対応できます。

---

## 現在適用されている文字列パッチ (Active String Patches)

プロンプトパッチに加えて、`agy-wrapper` は同じ長さのバイナリ文字列をインプレースで置換します：

- **`disable-play-logging`**：すべての `https://play.googleapis.com/log` を `http://0.0.0.0/logging-disabled`（各 31 バイト）に置換し、テレメトリ/ロギングの送信を無効化します。対象文字列がすでにパッチ済み、またはビルド内に存在しない場合は安全にスキップされます。

---

## プロジェクト構成

```
├── i18n/                 # 多言語ドキュメント (vi, kr, cn, ja)
├── src/                  # 機能別にモジュール化された Go ソースコード
│   ├── env/              # PATH 環境変数管理およびバイナリ検出
│   │   └── env.go
│   ├── meta/             # メタデータキャッシュおよび整合性検証
│   │   └── meta.go
│   ├── parser/           # バイナリ構造定義およびパーサー (PE, ELF, Mach-O)
│   │   ├── binary.go
│   │   ├── pe.go
│   │   ├── elf.go
│   │   └── macho.go
│   ├── patcher/          # バイナリパッチエンジンおよびメモリ再配置
│   │   └── patcher.go
│   ├── prompt/           # 埋め込みプロンプト管理および抽出エンジン
│   │   ├── prompts/      # 埋め込みプロンプト Markdown 定義
│   │   └── prompt.go
│   └── main.go           # CLI エントリポイント、引数ルーティング、プロキシ
├── go.mod                # Go モジュール仕様書
├── LICENSE               # MIT ライセンス
└── README.md             # 英語メインドキュメント
```

---

## 事前ビルド済みリリース (Releases)

バージョンタグ（例: `v1.0.0`）がプッシュされると、GitHub Releases に事前コンパイルされたバイナリ ZIP アーカイブが自動的に作成・公開されます：

| プラットフォーム | ZIP アーカイブ名 | 内包バイナリ |
| :--- | :--- | :--- |
| **Windows x86_64** | `agy-wrapper-v{VERSION}-windows-amd64.zip` | `agy.exe` |
| **Linux x86_64** | `agy-wrapper-v{VERSION}-linux-amd64.zip` | `agy` |
| **Linux ARM64** | `agy-wrapper-v{VERSION}-linux-arm64.zip` | `agy` |
| **macOS Intel** | `agy-wrapper-v{VERSION}-darwin-amd64.zip` | `agy` |
| **macOS Apple Silicon** | `agy-wrapper-v{VERSION}-darwin-arm64.zip` | `agy` |
| **macOS Universal** | `agy-wrapper-v{VERSION}-darwin-universal.zip` | `agy` |

*各 ZIP アーカイブには、実行可能ファイル、`LICENSE`、`README.md`、および `i18n/` 多言語ドキュメントフォルダが含まれます。*

---

## ビルド方法

ビルド出力先はすべて `dist/` ディレクトリです（Git 管理対象外）。

必要環境：**Go 1.21+**。

### Windows

**Windows ネイティブ（PowerShell）：**
```powershell
go build -ldflags="-s -w" -o dist/agy.exe ./src
```

**macOS/Linux から Windows 向けクロスコンパイル：**
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy.exe ./src
```

---

### Linux

**Linux ネイティブまたはクロスコンパイル（x86_64）：**
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy ./src
```

**ARM64（Raspberry Pi、AWS Graviton など）：**
```bash
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy ./src
```

---

### macOS

**Apple Silicon（M1/M2/M3/M4）：**
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy ./src
```

**Intel Mac（x86_64）：**
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy ./src
```

**ユニバーサルバイナリ（Apple Silicon + Intel 両対応）：**
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/agy-arm64 ./src
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/agy-amd64 ./src
lipo -create -output dist/agy dist/agy-arm64 dist/agy-amd64
rm dist/agy-arm64 dist/agy-amd64
```

---

## プロンプト不一致時の安全対策

`agy-wrapper` の実行時、対象バイナリの検査が行われます：
1. **正確なチェックサム一致**：対象バイナリ内のプロンプトが想定した元テキストと 1 バイト単位で完全一致する場合のみパッチを適用します。
2. **上流プロンプトが変更されていた場合**：
   - `Checksum mismatch — upstream prompt has changed` または `Prompt not found in binary` のエラーを表示し直ちに処理を中止します。
   - **元バイナリは一切変更・破損されません**（事前検証はすべてメモリ上で実施されます）。
   - ディスクへの書き込み前に自動バックアップ（`*.original`）が必ず作成されます。
3. **プロンプトの更新方法**：
   - プロンプトは組み込みリソースよりもディスク（`prompts/<platform>/` または `src/prompts/<platform>/`）のパスを優先的に読み込みます。
   - 上流の新バージョンがリリースされた際は、抽出した最新プロンプトを `prompts/` に配置するか再ビルドすることで即座に対応可能です。

---

## 使い方

1. `dist/` に生成された実行可能ファイルを任意のフォルダへ配置します。
2. 初回起動で PATH 自動設定およびドライラン検証を実行します：
   ```bash
   agy
   ```
3. ターミナルを再起動して更新された PATH を反映させ、通常の `agy` コマンドとして利用します：
   ```bash
   agy --help
   ```

### ドライラン（Dry-Run）検証

ディスク書き込みを行わずにバイナリ状態とパッチ適用のシミュレーションを行います：
```bash
agy --dry-run
```

### `agy` からのプロンプト抽出

バイナリに埋め込まれているシステムプロンプトを抽出・保存します：
```bash
# デフォルトで prompts/<platform>/ に全プロンプトを抽出
agy --extract-prompt

# 任意の出力ディレクトリを指定して抽出
agy --extract-prompt ./my_prompts

# 特定のツールのみ抽出（例：grill-me, plan, learn）
agy --extract-prompt --tool=grill-me

# 標準出力（stdout）へ直接出力
agy --extract-prompt --tool=grill-me --stdout
```

バックアップファイル（`.original`）が存在する場合は、上流のオリジナルプロンプト（`<tool>.original.md`）も自動で抽出されます。

---

## ライセンス

本プロジェクトは [MIT ライセンス](../../LICENSE) の下で公開されています。

# Copilot Instructions — image_getter

## プロジェクト概要

Go 製の画像一括ダウンローダー CLI ツール。指定 URL のページから `<img>` タグを収集し、`<a>` タグを辿って再帰的に画像をダウンロードする。

## 技術スタック

- **言語**: Go 1.26
- **モジュール**: `github.com/rssh-jp/image_getter`
- **主な依存**:
  - `github.com/PuerkitoBio/goquery` — HTML パース
  - `github.com/urfave/cli/v2` — CLI フレームワーク

## ディレクトリ構成と責務

| パス | 役割 |
|------|------|
| `cmd/image_getter/main.go` | エントリーポイント。CLI 定義・並行ダウンロード制御 |
| `imagegetter.go` | `ImageGetter` 構造体。URL から `<img>` を収集しチャネルに送信 |
| `imagesaver.go` | HTTP GET で画像を取得しローカルに保存する `SaveImage` 関数 |
| `config/config.go` | JSON 設定ファイルの読み込みパーサー（`Config` 構造体） |
| `imagegetter_test.go` | `ImageGetter` のユニットテスト（`httptest` 使用） |

## コーディング規約

- エラーは呼び出し元に返す。ロギングはエントリーポイント（`main.go`）近くで行う。
- `any` の使用は禁止。型安全性を維持する。
- 廃止済み API（`ioutil` など）は使用しない。`os.ReadFile` / `io` パッケージを使用する。
- 並行処理はセマフォ（バッファ付きチャネル）で最大並行数を制御する（現在: 8）。
- `goquery.NewDocument` は内部で HTTP 通信を行う点に注意（テスト時は `httptest.NewServer` を使用）。

## パッケージ依存方向

```
cmd/image_getter → imagegetter / imagesaver / config
imagegetter      → goquery
imagesaver       → net/http / os
config           → os / encoding/json
```

## 検証コマンド

```bash
go build ./...
go test ./...
go run ./cmd/image_getter -u <URL> -s <出力先> -d <深度>
```

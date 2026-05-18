# Copilot Instructions — image_getter

## プロジェクト概要

Go 製の画像一括ダウンローダー CLI ツール。指定 URL のページから `<img>` / `<picture source>` タグを収集し、`<a>` タグを **並列** に辿って再帰的に画像をダウンロードする。

## 技術スタック

- **言語**: Go 1.26
- **モジュール**: `github.com/rssh-jp/image_getter`
- **主な依存**:
  - `github.com/PuerkitoBio/goquery` — HTML パース
  - `github.com/urfave/cli/v2` — CLI フレームワーク

## ディレクトリ構成と責務

| パス | 役割 |
|------|------|
| `cmd/image_getter/main.go` | エントリーポイント。CLI 定義・consumer goroutine・ディレクトリリネーム |
| `internal/imagegetter/imagegetter.go` | `ImageGetter` 構造体。並列ページフェッチ・画像 URL 収集 |
| `internal/imagegetter/imagesaver.go` | HTTP GET で画像を取得しローカルに保存する `SaveImage` 関数 |
| `internal/imagegetter/imagegetter_test.go` | `ImageGetter` のユニットテスト（`httptest` 使用） |
| `internal/config/config.go` | JSON 設定ファイルの読み込みパーサー（`Config` 構造体） |

## 並行処理アーキテクチャ

```
Execute(urlStr, depth)
  └─ execute() × N goroutine  [fetchSem で最大16同時フェッチ]
       └─ seen map で訪問済み URL を重複排除
       └─ inst.URL (buffered chan, cap=128) に画像 URL を送信

consumer goroutine
  └─ for range inst.URL
       └─ SaveImage() × M goroutine  [semaphore で最大8同時保存]
            └─ seenDirs に保存先ディレクトリを記録

同期順序:
  inst.Execute() → inst.Close()  (全 execute goroutine 完了 → inst.URL close)
  → <-consumerDone               (全 SaveImage 完了)
  → renameWithCount()            (ファイル数確定後にリネーム)
```

## 主要な設計ポイント

- **ページ URL 重複排除**: `seen map[string]struct{}` + `seenMu sync.Mutex` で同一 URL を1回のみフェッチ
- **goroutine リーク防止**: `<a>` goroutine を生成する前に `g.wg.Add(1)` を呼ぶ（`Close()` の早期実行を防止）
- **ディレクトリリネーム**: 全ダウンロード完了後に `{ファイル数}_{元のディレクトリ名}` へ一括リネーム
- **ファイル名デコード**: `url.Parse` の `Path` フィールドを使用し、パーセントエンコードを復号（日本語ファイル名対応）
- **画像属性フォールバック**: `src` → `data-src` → `data-lazy-src` → `data-original` → `data-lazy` → `srcset` の順に探索
- **HTTP クライアント共有**: `userAgent` 定数と `httpClient` 変数をページフェッチ・画像保存で共用

## コーディング規約

- エラーは呼び出し元に返す。ロギングはエントリーポイント（`main.go`）近くで行う。
- `any` の使用は禁止。型安全性を維持する。
- 廃止済み API（`ioutil` など）は使用しない。`os.ReadFile` / `io` パッケージを使用する。
- goroutine 内で `log.Fatal` を使わない（`os.Exit` が defer を無視するため `log.Println` を使う）。
- `goquery.NewDocumentFromReader` を使用する（`NewDocument` は非推奨・内部で HTTP 通信）。テスト時は `httptest.NewServer` を使用。

## パッケージ依存方向

```
cmd/image_getter → internal/imagegetter / internal/config
internal/imagegetter → goquery
internal/config  → os / encoding/json
```

## 検証コマンド

```bash
make build   # go build -o bin/image_getter ./cmd/image_getter
make test    # go test -race ./...
make run URL=<URL> STORAGE=<出力先> DEPTH=<深度>
```

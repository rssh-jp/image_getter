# image_getter

指定した URL から画像を再帰的にダウンロードするCLIツール。

## 機能

- 指定URLページ内の `<img>` タグの画像を一括ダウンロード
- `<a>` タグを辿って指定深度まで再帰的に収集（`--depth` オプション）
- セマフォによる並行ダウンロード（最大8並行）
- ホスト名・パス構造を保持したローカルディレクトリへの保存

## 使い方

```bash
# ヘルプ表示
bin/image_getter -h

# 基本的な使い方（深度0 = 指定URLのみ）
bin/image_getter -s ./download -u https://example.com/ -d 0

# 深度1: 指定URLからリンクを1段階辿って収集
bin/image_getter -s ./download -u https://example.com/ -d 1
```

## パラメータ

| 省略名 | 名前 | 必須 | 説明 |
| ------ | ---- | ---- | ---- |
| `-u` | `--url` | ✅ | ダウンロード元のURL |
| `-s` | `--storage_path` | ✅ | ダウンロード先のローカルパス |
| `-d` | `--depth` | - | `<a>` タグを辿る深度（デフォルト: 0） |
| `-h` | `--help` | - | ヘルプ表示 |

## ビルド

```bash
go build -o bin/image_getter ./cmd/image_getter
```

## テスト

```bash
go test ./...
```

## 技術スタック

- Go 1.26
- [goquery](https://github.com/PuerkitoBio/goquery) — HTML パース・スクレイピング
- [urfave/cli](https://github.com/urfave/cli) — CLI フレームワーク

## ディレクトリ構成

```
.
├── cmd/image_getter/main.go   # エントリーポイント
├── config/config.go           # JSON 設定ファイルパーサー
├── imagegetter.go             # 画像URL収集ロジック
├── imagegetter_test.go        # テスト
├── imagesaver.go              # 画像ダウンロード・保存ロジック
└── bin/                       # ビルド成果物
```


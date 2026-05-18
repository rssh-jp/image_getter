# image_getter

指定した URL から画像を再帰的にダウンロードするCLIツール。

## 機能

- 指定URLページ内の `<img>` / `<picture source>` タグの画像を一括ダウンロード
- 遅延読み込み属性（`data-src`, `data-lazy-src`, `data-original`, `data-lazy`）にも対応
- `<a>` タグを辿って指定深度まで **並列** 再帰収集（最大16ページ同時フェッチ）
- 訪問済みページの重複排除（同一URLは1回のみフェッチ）
- セマフォによる並行ダウンロード（最大8並行）
- ホスト名・パス構造を保持したローカルディレクトリへの保存
- ダウンロード完了後、ディレクトリ名を `{ファイル数}_{元のディレクトリ名}` にリネーム
- URL パーセントエンコードを復号したファイル名で保存（日本語ファイル名対応）

## 使い方

```bash
# ヘルプ表示
bin/image_getter -h

# 基本的な使い方（深度0 = 指定URLのみ）
bin/image_getter -s ./download -u https://example.com/ -d 0

# 深度1: 指定URLからリンクを1段階辿って収集
bin/image_getter -s ./download -u https://example.com/ -d 1

# make 経由で実行（STORAGE のデフォルト: ./download、DEPTH のデフォルト: 0）
make run URL=https://example.com/ STORAGE=./download DEPTH=1
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
make build
# または
go build -o bin/image_getter ./cmd/image_getter
```

## テスト

```bash
make test
# または
go test ./...
```

## 技術スタック

- Go 1.26
- [goquery](https://github.com/PuerkitoBio/goquery) — HTML パース・スクレイピング
- [urfave/cli](https://github.com/urfave/cli) — CLI フレームワーク

## ディレクトリ構成

```
.
├── cmd/image_getter/main.go               # エントリーポイント
├── internal/
│   ├── imagegetter/
│   │   ├── imagegetter.go                 # 画像URL収集・並列ページフェッチ
│   │   ├── imagegetter_test.go            # テスト
│   │   └── imagesaver.go                 # 画像ダウンロード・保存
│   └── config/
│       └── config.go                      # JSON 設定ファイルパーサー
├── Makefile
└── bin/                                   # ビルド成果物
```


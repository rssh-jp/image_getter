# image_getter
- 画像ダウンローダー
- 用途は好き好きで、何か要望あればissueください。例) google検索して画像をダウンロードするようにして

# 使い方
```
bin/image_getter -h

bin/image_getter -s ./download -u https://golang.org/pkg/ -d 0
```

# パラメータ
| -s, --storage_path | ダウンロード先のローカルストレージパス |
| -u, --url | ダウンロード元のURL |
| -d, --depth | ダウンロード元のURLのaタグを見て深堀りしていく数。デフォルト0  |
| -h, --help | ヘルプ表示  |

# 注意事項
- ソースを使う場合はまだテストコードを書いてないことに注意してください。
- これからテストコードを追加していこうと思っています。


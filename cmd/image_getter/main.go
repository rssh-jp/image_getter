package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	imagegetter "github.com/rssh-jp/image_getter/internal/imagegetter"

	"github.com/urfave/cli/v2"
)

const (
	semaphore = 8
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Required: true,
				Usage:    "url",
			},
			&cli.StringFlag{
				Name:     "storage_path",
				Aliases:  []string{"s"},
				Required: true,
				Usage:    "download path",
			},
			&cli.IntFlag{
				Name:    "depth",
				Aliases: []string{"d"},
				Value:   0,
				Usage:   "url pursue depth",
			},
		},
		Action: func(c *cli.Context) error {
			log.Println("START")
			defer log.Println("END")

			confURL := c.String("url")
			confStoragePath := c.String("storage_path")
			confDepth := c.Int("depth")

			inst := imagegetter.New()

			var (
				wgRead     sync.WaitGroup
				mapURL     = make(map[string]struct{})
				mapURLMu   sync.Mutex
				seenDirs   = make(map[string]struct{})
				seenDirsMu sync.Mutex
			)

			// consumerDone は consumer goroutine が全 SaveImage を含めて完了した時点で close される。
			// renameWithCount はこの後にのみ実行する。
			consumerDone := make(chan struct{})
			go func() {
				defer close(consumerDone)
				ch := make(chan struct{}, semaphore)
				for data := range inst.URL {
					wgRead.Add(1)
					ch <- struct{}{}
					go func(url string) {
						defer func() {
							<-ch
							wgRead.Done()
						}()

						mapURLMu.Lock()
						if _, ok := mapURL[url]; ok {
							mapURLMu.Unlock()
							return
						}
						mapURL[url] = struct{}{}
						mapURLMu.Unlock()

						dir := getDir(url, confStoragePath)
						if err := imagegetter.SaveImage(url, dir); err != nil {
							log.Println(err)
							return
						}

						seenDirsMu.Lock()
						seenDirs[dir] = struct{}{}
						seenDirsMu.Unlock()
					}(data.SrcURL)
				}
				// inst.URL が close された後、残存 SaveImage goroutine の完了を待つ。
				wgRead.Wait()
			}()

			// Execute は同期呼び出し。内部 goroutine はまだ動いている場合がある。
			if err := inst.Execute(confURL, confDepth); err != nil {
				log.Println(err)
			}
			// Close は全 execute goroutine の完了を待ってから inst.URL を close する。
			inst.Close()

			// consumer goroutine（全 SaveImage 含む）の完了を待つ。
			<-consumerDone

			// 全ダウンロード完了後にディレクトリを一括リネーム
			for dir := range seenDirs {
				if err := renameWithCount(dir); err != nil {
					log.Println("rename:", err)
				}
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getDir(u, destpath string) string {
	workURL, err := url.Parse(u)
	if err != nil {
		return ""
	}

	list := strings.Split(strings.Trim(workURL.Path, "/"), "/")
	str := strings.Join(list[:len(list)-1], "_")
	dir := filepath.Join(destpath, workURL.Host, str)
	return dir
}

// renameWithCount はディレクトリ内のファイル数を数え、"{count}_{basename}" にリネームする。
// 全ダウンロード完了後に一度だけ呼ぶことを想定している。
func renameWithCount(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	if count == 0 {
		return nil
	}
	parent := filepath.Dir(dir)
	base := filepath.Base(dir)
	newDir := filepath.Join(parent, fmt.Sprintf("%d_%s", count, base))
	if dir == newDir {
		return nil
	}
	return os.Rename(dir, newDir)
}

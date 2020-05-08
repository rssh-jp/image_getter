package main

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rssh-jp/image_getter"

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
			defer inst.Close()

			var wg sync.WaitGroup
			var wgRead sync.WaitGroup

			wg.Add(1)

			mapURL := make(map[string]struct{})

			go func() {
				ch := make(chan struct{}, semaphore)
				for {
					select {
					case data := <-inst.URL:
						wgRead.Add(1)
						ch <- struct{}{}

						go func(url string) {
							defer func() {
								<-ch
								wgRead.Done()
							}()

							if _, ok := mapURL[url]; ok {
								return
							}

							err := imagegetter.SaveImage(url, getDir(url, confStoragePath))
							if err != nil {
								log.Fatal(err)
							}

							mapURL[url] = struct{}{}
						}(data.SrcURL)
					}
				}
			}()

			go func() {
				defer wg.Done()

				err := inst.Execute(confURL, confDepth)
				if err != nil {
					log.Fatal(err)
				}
			}()

			wait(inst, &wg, &wgRead)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func wait(i *imagegetter.ImageGetter, wg, wgRead *sync.WaitGroup) {
	wg.Wait()

	time.Sleep(time.Millisecond)

	wgRead.Wait()
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

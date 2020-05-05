package main

import (
	"flag"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rssh-jp/image_getter"
	"github.com/rssh-jp/image_getter/config"
)

var (
	conf   *config.Config
	mapURL = make(map[string]struct{})
)

func preprocess() error {
	var path string

	flag.StringVar(&path, "path", "config.json", "config.json path")
	flag.Parse()

	var err error

	conf, err = config.New(path)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Println("START")
	defer log.Println("END")

	err := preprocess()
	if err != nil {
		log.Fatal(err)
	}

	inst := imagegetter.New()
	defer inst.Close()

	var wg sync.WaitGroup
	var wgRead sync.WaitGroup

	wg.Add(1)

	go func() {
		for {
			select {
			case url := <-inst.URL:
				wgRead.Add(1)

				if _, ok := mapURL[url]; ok {
					continue
				}

				err := imagegetter.SaveImage(url, getDir(url, conf.StoragePath))
				if err != nil {
					log.Fatal(err)
				}

				mapURL[url] = struct{}{}
				wgRead.Done()
			}
		}
	}()

	go func() {
		defer wg.Done()

		err := inst.Execute(conf.Url, conf.Depth)
		if err != nil {
			log.Fatal(err)
		}
	}()

	wait(inst, &wg, &wgRead)
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

package imagegetter

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type ImageGetter struct {
	URL chan string
	wg  sync.WaitGroup
}

func New() *ImageGetter {
	return &ImageGetter{
		URL: make(chan string),
	}
}

func (g *ImageGetter) Close() {
	g.wg.Wait()

	close(g.URL)
}

func (g *ImageGetter) Execute(urlStr string, depth int) error {
	g.wg.Add(1)

	defer g.wg.Done()

	res, err := goquery.NewDocument(urlStr)
	if err != nil {
		return err
	}

	netUrl, _ := url.Parse(urlStr)
	fmt.Println(netUrl)
	res.Find("img").Each(func(_ int, s *goquery.Selection) {
		urlStr, _ := s.Attr("src")

		var workURL *url.URL
		workURL, err = netUrl.Parse(urlStr)
		if err != nil {
			return
		}

		g.URL <- workURL.String()
	})

	if err != nil {
		return err
	}

	if depth <= 0 {
		return nil
	}

	res.Find("a").Each(func(_ int, s *goquery.Selection) {
		wkUrl, _ := s.Attr("href")
		workURL, _ := netUrl.Parse(wkUrl)
		err = g.Execute(workURL.String(), depth-1)
		if err != nil {
			return
		}
	})

	if err != nil {
		return err
	}

	return nil
}

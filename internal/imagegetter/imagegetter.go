package imagegetter

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// userAgent は HTTP リクエストに付与する User-Agent 文字列。
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

const (
	// maxConcurrentFetches は同時ページフェッチ数の上限。
	maxConcurrentFetches = 16
	// urlChanBuffer は画像 URL チャネルのバッファサイズ。
	urlChanBuffer = 128
)

// lazySrcAttrs は画像 URL を探す属性名の優先順リスト。
// 遅延読み込みライブラリが使う data-src 系を src の次に確認する。
var lazySrcAttrs = []string{"src", "data-src", "data-lazy-src", "data-original", "data-lazy"}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}

type Data struct {
	BaseURL string
	SrcURL  string
}

type ImageGetter struct {
	URL      chan Data
	wg       sync.WaitGroup
	fetchSem chan struct{}       // 同時フェッチ数を制限するセマフォ
	seen     map[string]struct{} // 訪問済みページ URL の重複排除
	seenMu   sync.Mutex
}

func New() *ImageGetter {
	return &ImageGetter{
		URL:      make(chan Data, urlChanBuffer),
		fetchSem: make(chan struct{}, maxConcurrentFetches),
		seen:     make(map[string]struct{}),
	}
}

func (g *ImageGetter) Close() {
	g.wg.Wait()

	close(g.URL)
}

func fetchDocument(urlStr string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}

// Execute は urlStr のページを解析して画像 URL をチャネルに送信する。
// depth > 0 の場合は <a> タグを並列に辿る。
func (g *ImageGetter) Execute(urlStr string, depth int) error {
	g.wg.Add(1)
	defer g.wg.Done()

	return g.execute(urlStr, depth)
}

// execute は実際のフェッチ・解析・再帰処理を行う内部メソッド。
func (g *ImageGetter) execute(urlStr string, depth int) error {
	// 訪問済みページは再フェッチしない。
	g.seenMu.Lock()
	if _, ok := g.seen[urlStr]; ok {
		g.seenMu.Unlock()
		return nil
	}
	g.seen[urlStr] = struct{}{}
	g.seenMu.Unlock()

	// 同時フェッチ数を制限（フェッチ完了後すぐに解放）。
	g.fetchSem <- struct{}{}
	doc, err := fetchDocument(urlStr)
	<-g.fetchSem
	if err != nil {
		return err
	}

	baseURL, _ := url.Parse(urlStr)
	log.Println("EXECUTE URL:", baseURL)

	// <img> と <picture> 内の <source> から画像 URL を収集する。
	doc.Find("img, picture source").Each(func(_ int, s *goquery.Selection) {
		g.sendImageURLs(s, baseURL)
	})

	if depth <= 0 {
		return nil
	}

	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			return
		}
		linkURL, parseErr := baseURL.Parse(href)
		if parseErr != nil {
			return
		}
		// http/https 以外 (javascript:, mailto: 等) は辿らない。
		if linkURL.Scheme != "http" && linkURL.Scheme != "https" {
			return
		}
		// goroutine 生成前に wg.Add することで Close の早期実行を防ぐ。
		g.wg.Add(1)
		go func(u string) {
			defer g.wg.Done()
			if execErr := g.execute(u, depth-1); execErr != nil {
				log.Println("Execute error:", execErr)
			}
		}(linkURL.String())
	})

	return nil
}

// sendImageURLs は選択要素から画像 URL を抽出してチャネルに送信する。
// src → data-src 系の順に確認し、いずれも無効な場合は srcset にフォールバックする。
func (g *ImageGetter) sendImageURLs(s *goquery.Selection, baseURL *url.URL) {
	sent := false
	for _, attr := range lazySrcAttrs {
		rawSrc, exists := s.Attr(attr)
		if !exists || rawSrc == "" || strings.HasPrefix(rawSrc, "data:") {
			continue
		}
		imgURL, err := baseURL.Parse(rawSrc)
		if err != nil || !isValidImageURL(imgURL) {
			continue
		}
		g.URL <- Data{BaseURL: baseURL.String(), SrcURL: imgURL.String()}
		sent = true
		break
	}

	if !sent {
		if srcset, exists := s.Attr("srcset"); exists && srcset != "" {
			if imgURL := parseBestSrcset(srcset, baseURL); imgURL != nil {
				g.URL <- Data{BaseURL: baseURL.String(), SrcURL: imgURL.String()}
			}
		}
	}
}

// parseBestSrcset は srcset 属性値から最後（最高解像度）の URL を返す。
func parseBestSrcset(srcset string, baseURL *url.URL) *url.URL {
	var best *url.URL
	for _, part := range strings.Split(srcset, ",") {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) == 0 {
			continue
		}
		u, err := baseURL.Parse(fields[0])
		if err != nil || !isValidImageURL(u) {
			continue
		}
		best = u
	}
	return best
}

// isValidImageURL は http/https スキームで Host が存在する URL かどうかを返す。
func isValidImageURL(u *url.URL) bool {
	return u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}

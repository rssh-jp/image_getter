package imagegetter

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

func SaveImage(rawURL, dir string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	// parsed.Path はパーセントエンコードが復号された文字列。
	// 日本語などのマルチバイト文字もそのままファイル名として使える。
	name := path.Base(parsed.Path)
	if name == "" || name == "." || name == "/" {
		return nil
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	r, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}

	defer r.Body.Close()

	path := filepath.Join(dir, name)

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Println("os.MkdirAll err. : ", err)
		return err
	}

	err = copyToFile(path, r.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("DONE COPY. %s -> %s\n", rawURL, path)

	return nil
}

func copyToFile(name string, src io.Reader) (err error) {
	file, err := os.Create(name)
	if err != nil {
		log.Println("os.Create err. : ", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, src)
	if err != nil {
		log.Println("io.Copy err. : ", err)
		return
	}
	return
}

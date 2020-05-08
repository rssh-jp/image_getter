package imagegetter

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SaveImage(url, dir string) error {
	l := strings.Split(strings.Split(url, "?")[0], "/")

	name := l[len(l)-1]

	r, err := http.Get(url)
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

	err = copy(path, r.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("DONE COPY. %s -> %s\n", url, path)

	return nil
}

func copy(name string, src io.Reader) (err error) {
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

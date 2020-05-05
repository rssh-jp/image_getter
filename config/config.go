package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
)

// コンフィグ
type Config struct {
	Url         string `json:"url"`
	StoragePath string `json:"storage_path"`
	Depth       int    `json:"depth"`
}

func New(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := new(Config)

	dec := json.NewDecoder(strings.NewReader(string(buf)))

	for {
		if err := dec.Decode(conf); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}

	return conf, nil
}

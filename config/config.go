package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Locale  string            `toml:"locale"`
	Dest    string            `toml:"destination"`
	Network Network           `toml:"network"`
	Locales map[string]Locale `toml:"locales"`
}

type Network struct {
	Addr string `toml:"address"`
}

type Locale struct {
	Alias string `toml:"alias"`
	Dest  string `toml:"destination"`
}

func Parse(data []byte) (Config, error) {
	var c Config

	err := toml.Unmarshal(data, &c)
	if err != nil {
		return Config{}, err
	}

	if c.Locale == "" || c.Dest == "" {
		return Config{}, errors.New("Default values are not defined")
	}

	return c, nil
}

func Locate() string {
	name := "relocale.toml"

	paths := []string{
		filepath.Join("/etc/relocale", name),
		name,
	}

	found := ""
	for _, path := range paths {
		_, err := os.Stat(path)
		if err != nil {
			continue
		}

		found = path
	}

	return found
}

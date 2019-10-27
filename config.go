package torbula

import (
	"fmt"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// Config is parsed result of the ini file
type config struct {
	PathSrc string
	PathTmp string
	PathDst string
	PathLog string
}

var defaultConfig config

// ParseConfig parses ini file into cfg.
func parseConfig(file string) error {
	loaded, err := ini.Load(file)
	if err != nil {
		return err
	}
	section := loaded.Section("")
	if section == nil {
		return fmt.Errorf("section not found")
	}

	getPath := func(key string) (string, error) {
		value := section.Key(key)
		if value == nil {
			return "", fmt.Errorf("%q not found from config file", key)
		}
		path := value.String()
		if filepath.IsAbs(path) {
			return "", fmt.Errorf("%s=%q, the path should be relative", key, path)
		}
		return path, nil
	}

	var config = &defaultConfig
	config.PathSrc, err = getPath("src_dir")
	if err != nil {
		return err
	}
	config.PathTmp, err = getPath("tmp_dir")
	if err != nil {
		return err
	}
	config.PathDst, err = getPath("dst_dir")
	if err != nil {
		return err
	}
	config.PathLog, err = getPath("log_dir")
	if err != nil {
		return err
	}
	return nil
}

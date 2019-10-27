package torbula

import (
	"fmt"
	"path/filepath"
	"time"

	"gopkg.in/ini.v1"
)

// Config is parsed result of the ini file
type config struct {
	pathSrc  string
	pathTmp  string
	pathDst  string
	pathLog  string
	seedTime time.Duration
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
			return path, nil
		}
		path, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		return path, nil
	}

	getDuration := func(key string, def time.Duration) (dur time.Duration, err error) {
		if !section.HasKey(key) {
			return def, nil
		}
		value, err := section.GetKey(key)
		if err != nil {
			return dur, err
		}
		dur, err = time.ParseDuration(value.String())
		if err != nil {
			return def, err
		}
		return dur, nil
	}

	var config = &defaultConfig
	config.pathSrc, err = getPath("src_dir")
	if err != nil {
		return err
	}
	config.pathTmp, err = getPath("tmp_dir")
	if err != nil {
		return err
	}
	config.pathDst, err = getPath("dst_dir")
	if err != nil {
		return err
	}
	config.pathLog, err = getPath("log_dir")
	if err != nil {
		return err
	}
	config.seedTime, err = getDuration("seed_time", 6*time.Hour)
	if err != nil {
		return err
	}
	return nil
}

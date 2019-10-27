package torbula

import (
	"path/filepath"
	"strings"
)

type progress struct {
	state
	path string
}

type state int

const (
	invalid state = iota
	detected
	downloading
	seeding
	seedend
)

func isTorrent(file string) bool {
	return strings.ToLower(filepath.Ext(file)) == ".torrent"
}

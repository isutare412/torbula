package torbula

import (
	"path/filepath"
	"strings"
	"sync/atomic"
)

type progress struct {
	id uint64
	state
	path string
}

var idCount *uint64 = new(uint64)

type state int

const (
	invalid state = iota
	detected
	downloading
	seeding
	seedend
)

func newProgress() *progress {
	return &progress{id: atomic.AddUint64(idCount, 1)}
}

func isTorrent(file string) bool {
	return strings.ToLower(filepath.Ext(file)) == ".torrent"
}

package torbula

import (
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/anacrolix/torrent"
)

var idCount *uint64 = new(uint64)

type progID uint64
type state int

const (
	invalid state = iota
	detected
	downloading
	seeding
	seedend
)

type progress struct {
	progID
	state
	path string
	hash torrent.InfoHash
}

func newProgress() *progress {
	return &progress{progID: progID(atomic.AddUint64(idCount, 1))}
}

type progressPool struct {
	sync.Mutex
	pros map[progID]*progress
}

func (pp *progressPool) newProgress(path string) (progID, bool) {
	if pp.hasPath(path) {
		return 0, false
	}
	pp.Lock()
	defer pp.Unlock()
	p := newProgress()
	p.path = path
	pp.pros[p.progID] = p
	return p.progID, true
}

func (pp *progressPool) hasID(id progID) bool {
	pp.Lock()
	defer pp.Unlock()
	if _, ok := pp.pros[id]; ok {
		return true
	}
	return false
}

func (pp *progressPool) hasPath(path string) bool {
	pp.Lock()
	defer pp.Unlock()
	for _, p := range pp.pros {
		if p.path == path {
			return true
		}
	}
	return false
}

func (pp *progressPool) path(id progID) (string, bool) {
	if !pp.hasID(id) {
		return "", false
	}
	pp.Lock()
	defer pp.Unlock()
	return pp.pros[id].path, true
}

func (pp *progressPool) progress(id progID) (progress, bool) {
	if !pp.hasID(id) {
		return progress{}, false
	}
	pp.Lock()
	defer pp.Unlock()
	return *pp.pros[id], true
}

func (pp *progressPool) setState(id progID, s state) bool {
	if !pp.hasID(id) {
		return false
	}
	pp.Lock()
	defer pp.Unlock()
	pp.pros[id].state = s
	return true
}

func (pp *progressPool) setHash(id progID, h torrent.InfoHash) bool {
	if !pp.hasID(id) {
		return false
	}
	pp.Lock()
	defer pp.Unlock()
	pp.pros[id].hash = h
	return true
}

func isTorrent(file string) bool {
	return strings.ToLower(filepath.Ext(file)) == ".torrent"
}

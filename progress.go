package torbula

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

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

func (s state) String() string {
	return [...]string{
		"invalid",
		"detected",
		"downloading",
		"seeding",
		"completed",
	}[s]
}

type progress struct {
	progID
	state
	path string
	name string
	hash torrent.InfoHash
	size int64

	start    time.Time
	end      time.Time
	finished bool
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

func (pp *progressPool) erase(id progID) bool {
	pp.Lock()
	defer pp.Unlock()
	if _, ok := pp.pros[id]; !ok {
		return false
	}
	delete(pp.pros, id)
	return true
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

func (pp *progressPool) finished(id progID) bool {
	if !pp.hasID(id) {
		return false
	}
	pp.Lock()
	defer pp.Unlock()
	return pp.pros[id].finished
}

func (pp *progressPool) path(id progID) (string, bool) {
	if !pp.hasID(id) {
		return "", false
	}
	pp.Lock()
	defer pp.Unlock()
	return pp.pros[id].path, true
}

func (pp *progressPool) hash(id progID) (torrent.InfoHash, bool) {
	if !pp.hasID(id) {
		return torrent.InfoHash{}, false
	}
	pp.Lock()
	defer pp.Unlock()
	return pp.pros[id].hash, true
}

func (pp *progressPool) sinceEnd(id progID) (time.Duration, bool) {
	if !pp.finished(id) {
		return 0, false
	}
	return time.Since(pp.pros[id].end), true
}

func (pp *progressPool) findByHash(h torrent.InfoHash) (progID, bool) {
	pp.Lock()
	defer pp.Unlock()
	for _, p := range pp.pros {
		if p.hash == h {
			return p.progID, true
		}
	}
	return 0, false
}

func (pp *progressPool) progress(id progID) (progress, bool) {
	if !pp.hasID(id) {
		return progress{}, false
	}
	pp.Lock()
	defer pp.Unlock()
	return *pp.pros[id], true
}

func (pp *progressPool) forEach(cb func(*progress)) {
	pp.Lock()
	defer pp.Unlock()

	ids := make([]progID, 0, len(pp.pros))
	for id := range pp.pros {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	for _, id := range ids {
		cb(pp.pros[id])
	}
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

func (pp *progressPool) setInfo(id progID, t *torrent.Torrent) bool {
	if !pp.hasID(id) || t == nil {
		return false
	}
	pp.Lock()
	defer pp.Unlock()
	pp.pros[id].name = t.Name()
	pp.pros[id].hash = t.InfoHash()
	pp.pros[id].size = t.Length()
	return true
}

func (pp *progressPool) setStart(id progID) {
	if !pp.hasID(id) {
		return
	}
	pp.Lock()
	defer pp.Unlock()
	pp.pros[id].start = time.Now()
}

func (pp *progressPool) setEnd(id progID) {
	if !pp.hasID(id) {
		return
	}
	pp.Lock()
	defer pp.Unlock()
	pp.pros[id].end = time.Now()
	pp.pros[id].finished = true
}

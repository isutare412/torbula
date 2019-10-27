package torbula

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
)

// Server automatically download torrents
type Server struct {
	// downloads torrents with torrentClient
	torrentClient *torrent.Client

	// absolute path. for internal usage.
	pathSrc  string
	pathTmp  string
	pathDst  string
	pathHome string

	// seeding time since download has completed.
	seedTime time.Duration

	mu      sync.Mutex
	onGoing map[uint64]*progress

	detected chan uint64
}

// Run start s and block. Run stops only if error occured.
func (s *Server) Run() error {
	defer s.torrentClient.Close()
	if err := s.ready(); err != nil {
		return err
	}

	go s.download()
	s.startDetect()
	return nil
}

func (s *Server) ready() error {
	if err := logInit(defaultConfig.pathLog); err != nil {
		return err
	}

	makedirs := func(paths ...string) error {
		for _, p := range paths {
			err := os.MkdirAll(p, 0755)
			if err != nil {
				return fmt.Errorf("failed create %q dir: %v", p, err)
			}
		}
		return nil
	}
	if err := makedirs(s.pathSrc, s.pathTmp, s.pathDst); err != nil {
		return err
	}
	if err := os.Chdir(s.pathTmp); err != nil {
		return fmt.Errorf("failed ready: %s", err)
	}
	return nil
}

func (s *Server) startDetect() {
	tick := time.Tick(500 * time.Millisecond)
	for {
		select {
		case <-tick:
			err := s.detect()
			if err != nil {
				logWarning("%s", err)
				continue
			}
		}
	}
}

func (s *Server) detect() error {
	err := filepath.Walk(
		s.pathSrc,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !isTorrent(info.Name()) {
				return nil
			}

			if id, ok := s.Add(path); ok {
				logAlways("detected: %q", path)
				s.detected <- id
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("failed detect: %s", err)
	}
	return nil
}

func (s *Server) download() {
	for id := range s.detected {
		path, ok := s.Path(id)
		if !ok {
			logWarning("failed to download: id(%d) not found", id)
			continue
		}
		t, err := s.torrentClient.AddTorrentFromFile(path)
		if err != nil {
			logWarning("failed to download: %s", err)
			continue
		}
			t.DownloadAll()
			logAlways("start download: %q", t.Name())
	}
}

// Has checks path is managed by Server. Returns progress id if exists.
func (s *Server) Has(path string) (uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.onGoing {
		if p.path == path {
			return p.id, true
		}
	}
	return 0, false
}

// Path returns relative filepath of a progress with id.
func (s *Server) Path(id uint64) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.onGoing[id]; ok {
		return p.path, true
	}
	return "", false
}

// Add adds path to onGoing, which manages download status.
// Returns false if path already exists. Always returns progress id.
func (s *Server) Add(path string) (uint64, bool) {
	if id, ok := s.Has(path); ok {
		return id, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p := newProgress()
	p.state = detected
	p.path = path
	s.onGoing[p.id] = p
	return p.id, true
}

// NewServer create server instance from iniFile
func NewServer(iniFile string) (*Server, error) {
	err := parseConfig(iniFile)
	if err != nil {
		return nil, err
	}
	client, err := torrent.NewClient(nil)
	if err != nil {
		return nil, err
	}

	curDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var config = &defaultConfig
	return &Server{
		torrentClient: client,
		pathSrc:       config.pathSrc,
		pathTmp:       config.pathTmp,
		pathDst:       config.pathDst,
		pathHome:      curDir,
		seedTime:      config.seedTime,
		onGoing:       make(map[uint64]*progress),
		detected:      make(chan uint64),
	}, nil
}

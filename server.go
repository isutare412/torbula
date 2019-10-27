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

	// torrent files are read from pathSrc
	pathSrc string
	// torrents are downloaded into pathTmp
	pathTmp string
	// download completed files are moved from pathTmp to pathDst
	pathDst string

	mu      sync.Mutex
	onGoing map[uint64]*progress
}

// Run start s and block. Run stops only if error occured.
func (s *Server) Run() error {
	defer s.torrentClient.Close()

	if err := logInit(defaultConfig.PathLog); err != nil {
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

	s.startDetect()
	return nil
}

func (s *Server) startDetect() {
	tick := time.Tick(500 * time.Millisecond)
	for {
		select {
		case <-tick:
			s.detect()
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

			if s.Add(path) {
				logAlways("detected: %s", path)
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("failed detect: %s", err)
	}
	return nil
}

// Has checks path is managed by Server.
func (s *Server) Has(path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.onGoing {
		if p.path == path {
			return true
		}
	}
	return false
}

// Add adds path to onGoing, which manages download status.
func (s *Server) Add(path string) bool {
	if s.Has(path) {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p := newProgress()
	p.state = detected
	p.path = path
	s.onGoing[p.id] = p
	return true
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

	var config = &defaultConfig
	return &Server{
		torrentClient: client,
		pathSrc:       config.PathSrc,
		pathTmp:       config.PathTmp,
		pathDst:       config.PathDst,
		onGoing:       make(map[uint64]*progress),
	}, nil
}

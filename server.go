package torbula

import (
	"fmt"
	"os"
	"path/filepath"
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

	onGoing map[string]progress
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

	tick := time.Tick(500 * time.Millisecond)
	for {
		select {
		case <-tick:
			s.tick()
		}
	}
}

func (s *Server) tick() {
	if err := s.detect(); err != nil {
		logWarning("%s", err)
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
			if _, ok := s.onGoing[path]; ok {
				// already detected
				return nil
			}

			s.onGoing[path] = progress{state: detected, path: path}
			logAlways("detected: %s", path)
			return nil
		})
	if err != nil {
		return fmt.Errorf("failed detect: %s", err)
	}
	return nil
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
		onGoing:       make(map[string]progress),
	}, nil
}

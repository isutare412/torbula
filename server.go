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

	// absolute path. for internal usage.
	pathSrc  string
	pathTmp  string
	pathDst  string
	pathHome string

	// seeding time since download has completed.
	seedTime time.Duration

	pool     progressPool
	detected chan progID
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

			if id, ok := s.pool.newProgress(path); ok {
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
		path, ok := s.pool.path(id)
		if !ok {
			logWarning("failed to download: id(%d) not found", id)
			continue
		}
		t, err := s.torrentClient.AddTorrentFromFile(path)
		if err != nil {
			logWarning("failed to download: %s", err)
			continue
		}

		go func(id progID) {
			<-t.GotInfo()
			t.DownloadAll()
			s.pool.setHash(id, t.InfoHash())
			s.pool.setState(id, downloading)
			logAlways("start download: %q", t.Name())
		}(id)
	}
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
		detected:      make(chan progID),
		pool:          progressPool{pros: make(map[progID]*progress)},
	}, nil
}

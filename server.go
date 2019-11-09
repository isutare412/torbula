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
	todrop   chan torrent.InfoHash
}

// Run start Server and block. Run stops only if error occured.
func (s *Server) Run() error {
	defer s.torrentClient.Close()
	if err := s.ready(); err != nil {
		return err
	}

	go s.downloadStart()
	go s.downloadEnd()
	go s.drop()
	s.detect()
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

func (s *Server) detect() {
	for range time.Tick(1 * time.Second) {
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
					s.pool.setState(id, detected)
					s.detected <- id
				}
				return nil
			},
		)
		if err != nil {
			logWarning("%s", err)
			continue
		}
	}
}

func (s *Server) downloadStart() {
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
			s.pool.setStart(id)
			logAlways("start download: %v", t)
		}(id)
	}
}

func (s *Server) downloadEnd() {
	for range time.Tick(1 * time.Second) {
		var dropped []torrent.InfoHash
		for _, t := range s.torrentClient.Torrents() {
			if t.BytesCompleted() < t.Length() {
				continue
			}
			hash := t.InfoHash()
			id, ok := s.pool.findByHash(hash)
			if !ok {
				logWarning("downloadEnd: progress not found: %q", t)
				continue
			}
			if !s.pool.finished(id) {
				s.pool.setState(id, seeding)
				s.pool.setEnd(id)
			}

			duration, ok := s.pool.sinceEnd(id)
			if !ok {
				logWarning("downloadEnd: duration is invalid: %q", t)
				continue
			}
			if duration >= s.seedTime {
				s.pool.setState(id, seedend)
				dropped = append(dropped, hash)
			}
		}

		for _, h := range dropped {
			s.todrop <- h
		}
	}
}

func (s *Server) drop() {
	for hash := range s.todrop {
		id, ok := s.pool.findByHash(hash)
		if !ok {
			logWarning("drop: progress not found: hash(%v)", hash)
			continue
		}
		t, ok := s.torrentClient.Torrent(hash)
		if !ok {
			logWarning("drop: torrent not found: id(%d) hash(%v)", id, hash)
			continue
		}
		t.Drop()
		logAlways("dropped torrent: %q", t.Name())
	}
}

// NewServer create server instance from iniFile
func NewServer(iniFile string) (*Server, error) {
	err := parseConfig(iniFile)
	if err != nil {
		return nil, err
	}

	var config = &defaultConfig
	tconf := torrent.NewDefaultClientConfig()
	if config.seedTime > 0 {
		tconf.Seed = true
	}
	client, err := torrent.NewClient(tconf)
	if err != nil {
		return nil, err
	}

	curDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &Server{
		torrentClient: client,
		pathSrc:       config.pathSrc,
		pathTmp:       config.pathTmp,
		pathDst:       config.pathDst,
		pathHome:      curDir,
		seedTime:      config.seedTime,
		detected:      make(chan progID),
		todrop:        make(chan torrent.InfoHash),
		pool:          progressPool{pros: make(map[progID]*progress)},
	}, nil
}

package torbula

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kjk/dailyrotate"
)

var logger struct {
	path string
	file *dailyrotate.File
}

func logInit(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed create log dir: %s %v", path, err)
	}

	logger.path = path
	logger.file, err = dailyrotate.NewFileWithPathGenerator(
		func(t time.Time) string {
			time := t.Format("2006-01-02") + ".log"
			return filepath.Join(path, time)
		},
		nil,
	)
	if err != nil {
		return fmt.Errorf("init logger: %v", err)
	}
	logger.file.Location = time.Local
	return nil
}

func logTagged(tag string, format string, v ...interface{}) {
	var buf []byte
	builder := bytes.NewBuffer(buf)
	builder.WriteString(fmt.Sprintf("%s[%s] ", tag, time.Now().Format("15:04:05")))
	builder.WriteString(fmt.Sprintf(format, v...))
	builder.WriteRune('\n')
	logger.file.Write(builder.Bytes())
}

func logAlways(format string, v ...interface{}) {
	logTagged("[ALWAYS]", format, v...)
}

func logWarning(format string, v ...interface{}) {
	logTagged("[WARNING]", format, v...)
}

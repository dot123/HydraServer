package log

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Level         int
	Format        string
	Output        string
	OutputFile    string
	RotationCount int
	RotationTime  int
}

func InitLogger(c *Config, logger *logrus.Logger) (func(), error) {
	logger.SetLevel(logrus.Level(c.Level))
	switch c.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.DateTime})
	default:
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: time.DateTime})
	}

	var rl *rotatelogs.RotateLogs
	if c.Output != "" {
		switch c.Output {
		case "stdout":
			logger.SetOutput(os.Stdout)
		case "stderr":
			logger.SetOutput(os.Stderr)
		case "file":
			if name := c.OutputFile; name != "" {
				dir := filepath.Dir(name)
				ext := filepath.Ext(name)
				_ = os.MkdirAll(dir, 0777)
				r, err := rotatelogs.New(strings.TrimSuffix(name, ext)+"_%Y-%m-%d-%H-%M"+ext,
					rotatelogs.WithLinkName(name),
					rotatelogs.WithRotationTime(time.Duration(c.RotationTime)*time.Second),
					rotatelogs.WithRotationCount(uint(c.RotationCount)))
				if err != nil {
					return nil, err
				}

				logger.SetOutput(r)
				rl = r
			}
		}
	}

	return func() {
		if rl != nil {
			rl.Close()
		}
	}, nil
}

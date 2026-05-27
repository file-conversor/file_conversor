// internal/logger/setup.go

package logger

import (
	"fmt"
	"io"
	"log/slog"
)

var log *slog.Logger = slog.Default()

// SetupLogger initializes the logger with the provided configuration and returns an io.Closer for any resources that need to be closed (like file handles). It panics if there is an error during setup.
func SetupLogger(cfg *Config) (io.Closer, error) {
	var file io.Closer
	var err error
	log, file, err = cfg.New()
	if err != nil {
		if file != nil {
			file.Close()
		}
		return nil, fmt.Errorf("logger setup: %w\n", err)
	}
	slog.SetDefault(log)
	return file, nil
}

func Debugf(msg string, args ...any) {
	log.Debug(fmt.Sprintf(msg, args...))
}

func Infof(msg string, args ...any) {
	log.Info(fmt.Sprintf(msg, args...))
}

func Warnf(msg string, args ...any) {
	log.Warn(fmt.Sprintf(msg, args...))
}

func Errorf(msg string, args ...any) {
	log.Error(fmt.Sprintf(msg, args...))
}

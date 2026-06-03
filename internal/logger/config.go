// internal/logger/setup.go

package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Config struct {
	// Terminal
	TerminalLevel  Level
	TerminalFormat Format // pretty | text | json
	TerminalOutIo  io.Writer

	// File
	LogFile    string // "" = no file logging
	FileLevel  Level
	FileFormat Format // text | json recommended

	// AddSource adds file:line to every log entry — useful in debug builds
	AddSource bool
}

func DefaultConfig(logfile string) *Config {
	return &Config{
		TerminalLevel:  InfoLevel,    // less noise on terminal
		TerminalFormat: PrettyFormat, // human-friendly for terminal
		TerminalOutIo:  os.Stderr,    // log to stderr by default

		LogFile:    logfile,
		FileLevel:  DebugLevel, // capture everything to disk
		FileFormat: TextFormat, // structured logs are easier to analyze and query

		AddSource: false, // will not work here, as using custom functions for logging, not slog's built-in methods
	}
}

func (cfg *Config) New() (*slog.Logger, io.Closer, error) {
	var handlers []slog.Handler

	// --- Terminal handler ---
	termHandler := cfg.newHandler(cfg.TerminalOutIo, cfg.TerminalFormat, cfg.TerminalLevel)
	handlers = append(handlers, termHandler)

	// --- File handler ---
	var fileCloser io.Closer = io.NopCloser(nil)
	if cfg.LogFile != "" {
		f, err := os.OpenFile(
			cfg.LogFile,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0o644,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file: %w", err)
		}
		fileCloser = f

		fileHandler := cfg.newHandler(f, cfg.FileFormat, cfg.FileLevel)
		handlers = append(handlers, fileHandler)
	}

	logger := slog.New(NewFanoutHandler(handlers...))
	return logger, fileCloser, nil
}

func (cfg *Config) newHandler(out io.Writer, format Format, level Level) slog.Handler {
	opts := slog.HandlerOptions{
		Level:     level.Get(),
		AddSource: cfg.AddSource,
	}
	return format.GetHandler(out, opts)
}

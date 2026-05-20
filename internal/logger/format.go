// internal/logger/format.go

package logger

import (
	"io"
	"log/slog"
)

type Format string

const (
	JSONFormat   Format = "json"
	TextFormat   Format = "text"
	PrettyFormat Format = "pretty"
)

func (f *Format) GetHandler(out io.Writer, opts slog.HandlerOptions) slog.Handler {
	var handler slog.Handler
	switch *f {
	case JSONFormat:
		handler = slog.NewJSONHandler(out, &opts)
	case TextFormat:
		handler = slog.NewTextHandler(out, &opts)
	case PrettyFormat:
		handler = NewPrettyHandler(out, &opts)
	default:
		handler = NewPrettyHandler(out, &opts)
	}
	return handler
}

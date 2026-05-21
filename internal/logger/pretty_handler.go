// internal/logger/logger.go
package logger

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

// PrettyHandler is a human-readable handler with colour and aligned output.
// Wraps slog internals — no external dependencies.
type PrettyHandler struct {
	w      io.Writer
	opts   slog.HandlerOptions
	attrs  []slog.Attr
	groups []string
}

func NewPrettyHandler(w io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	return &PrettyHandler{w: w, opts: *opts}
}

func (h *PrettyHandler) Enabled(_ context.Context, l slog.Level) bool {
	min := h.opts.Level
	if min == nil {
		min = slog.LevelInfo
	}
	return l >= min.Level()
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	// level
	lvl := r.Level.String()

	// Build attr string
	var attrsStrBuilder strings.Builder
	attrsStrBuilder.Grow(64) // preallocate some space to avoid multiple allocations
	fn := func(a slog.Attr) bool {
		// faster to build string than concat or sprintf
		attrsStrBuilder.WriteByte(' ')
		attrsStrBuilder.WriteString(a.Key)
		attrsStrBuilder.WriteByte('=')
		attrsStrBuilder.WriteString(a.Value.String())
		return true
	}
	for _, a := range h.attrs {
		fn(a)
	}
	r.Attrs(fn)

	// allocate final log line in one go and write to output
	var logStrBuilder strings.Builder
	logStrBuilder.Grow(128 + attrsStrBuilder.Len())

	logStrBuilder.WriteByte('[')
	logStrBuilder.WriteString(lvl)
	logStrBuilder.WriteByte(']')
	if lvl == "INFO" {
		logStrBuilder.WriteByte(' ')
	}
	logStrBuilder.WriteString(" - ")
	logStrBuilder.WriteString(r.Message)
	logStrBuilder.WriteString(attrsStrBuilder.String())

	_, err := io.WriteString(h.w, logStrBuilder.String())
	return err
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newH := *h
	newH.attrs = append(h.attrs, attrs...)
	return &newH
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	newH := *h
	newH.groups = append(h.groups, name)
	return &newH
}

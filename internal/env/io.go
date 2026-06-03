// internal/env/io.go

package env

import (
	"io"
	"sync"
)

type IoProxy struct {
	mu sync.RWMutex
	w  io.Writer
}

func NewIoProxy(dest io.Writer) *IoProxy {
	return &IoProxy{
		w: dest,
	}
}

func (lp *IoProxy) Write(p []byte) (n int, err error) {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	return lp.w.Write(p)
}

func (lp *IoProxy) RouteTo(w io.Writer) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	lp.w = w
}

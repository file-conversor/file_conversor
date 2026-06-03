// internal/env/pipe.go

package env

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type PipeMgr struct {
	PipeReader *os.File
	PipeWriter *os.File

	wg      sync.WaitGroup
	errChan chan error

	lock    sync.Mutex
	started bool
	stopped bool
}

func NewPipeManager() (*PipeMgr, error) {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	return &PipeMgr{
		PipeReader: pipeReader,
		PipeWriter: pipeWriter,
		errChan:    make(chan error, 1), // buffered channel to avoid blocking
	}, nil
}

// Start begins copying data from the pipe to the destination writer in a separate goroutine.
func (pm *PipeMgr) Start(destination io.Writer) error {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	if destination == nil {
		return fmt.Errorf("pipe manager destination cannot be nil")
	}

	if pm.started || pm.stopped {
		return fmt.Errorf("pipe manager already started or stopped")
	}
	pm.started = true // set state to started (prevent multiple starts)

	// Multiplex the stream concurrently
	pm.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				pm.errChan <- fmt.Errorf("pipe stream panic: %v", r)
			}
			pm.wg.Done()
		}()

		// Copy data from the pipe to the destination (until EOF, pipe closed, or error)
		_, err := io.Copy(destination, pm.PipeReader)

		// io.EOF is the expected, natural termination of io.Copy when the writer closes.
		if err != nil {
			pm.errChan <- fmt.Errorf("pipe stream failed: %w", err)
		} else {
			pm.errChan <- nil // Signal success
		}
	}()
	return nil
}

// Stop signals the copying goroutine to stop by closing the pipe writer, waits for it to finish, and returns any error that occurred during copying or closing the pipe.
func (pm *PipeMgr) Stop() error {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	if pm.stopped {
		return nil // already stopped, nothing to do
	}
	pm.stopped = true // set state to stopped (prevent multiple stops)

	// wait for the goroutine to finish and capture any error it reports
	defer close(pm.errChan)
	errGrp := pm.PipeWriter.Close()
	if pm.started {
		// only wait for the goroutine if it was started
		// (avoid deadlock when Stop is called without Start)
		pm.wg.Wait()
		// get any error from the goroutine (nil if success)
		errGrp = errors.Join(errGrp, <-pm.errChan)
	}
	return errors.Join(errGrp, pm.PipeReader.Close())
}

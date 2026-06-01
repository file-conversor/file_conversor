// internal/env/progress_bar.go

package progress

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// ----------------------
// PROGRESS BAR STYLE
// ----------------------

type ProgressBarCfg struct {
	total   int64
	style   mpb.BarFillerBuilder
	options []mpb.BarOption
}

func NewBarCfg(name string, total int64, removeOnComplete bool) *ProgressBarCfg {
	progressBarCfg := &ProgressBarCfg{}

	const MAX_NAME_LEN = 20
	nameShort := name
	if len(name) > MAX_NAME_LEN {
		nameShort = name[:MAX_NAME_LEN-3] + "..." // truncate name if too long
	}

	prependDecor := []decor.Decorator{
		decor.Name(nameShort, decor.WC{
			W: len(nameShort),
			C: decor.DSyncWidthR,
		}),
	}
	appendDecor := []decor.Decorator{
		decor.OnComplete(
			decor.OnAbort(decor.Name(""), " failed"), "   done",
		),
		decor.Name(" | ET: "),
		decor.Elapsed(decor.ET_STYLE_MMSS, decor.WCSyncSpaceR),
	}

	if total <= 0 {
		progressBarCfg.total = -1 // indefinite mode
		progressBarCfg.style = defaultBouncingBarStyle()
	} else {
		progressBarCfg.total = total // definite mode
		progressBarCfg.style = defaultBarStyle()
		appendDecor = append(appendDecor,
			decor.OnCompleteOrOnAbort(decor.Name(" | ETA: "), ""),
			decor.OnCompleteOrOnAbort(decor.EwmaETA(decor.ET_STYLE_MMSS, 30, decor.WCSyncSpaceR), ""),
		)
	}

	progressBarCfg.options = append(progressBarCfg.options,
		mpb.PrependDecorators(prependDecor...),
		mpb.AppendDecorators(appendDecor...),
	)
	if removeOnComplete {
		progressBarCfg.options = append(progressBarCfg.options,
			mpb.BarRemoveOnComplete(),
		)
	}
	return progressBarCfg
}

func (c *ProgressBarCfg) IsIndefinite() bool {
	return c.total <= 0
}

// ----------------------
// PROGRESS BAR MANAGER
// ----------------------

type ProgressBarMgr struct {
	// controls the maximum number of concurrent bars
	maxWorkers  int
	workersChan chan struct{}

	// ensure thread safety when adding bars and waiting
	mu sync.Mutex
	wg *sync.WaitGroup

	// progress container from mpb library
	p        *mpb.Progress
	errChans []chan error
	finished bool
}

// NewProgressBarMgr creates a new ProgressBarMgr with the given style.
func NewProgressBarMgr(maxWorkers int, fps int) *ProgressBarMgr {
	var wg = &sync.WaitGroup{}
	var p = mpb.New(
		mpb.WithRefreshRate(time.Duration(1000/fps)*time.Millisecond), // set refresh rate (smooth animation)
		mpb.WithOutput(os.Stderr), // always write to stderr for progress bars
		mpb.WithWaitGroup(wg),     // progress manager waits for all bars to finish
	)
	var cpuCores = runtime.NumCPU()
	if maxWorkers <= 0 || maxWorkers > cpuCores*2 {
		maxWorkers = cpuCores * 2 // default to 2x CPU cores if invalid value provided
	}
	progressMgr := &ProgressBarMgr{
		maxWorkers:  maxWorkers,
		workersChan: make(chan struct{}, maxWorkers),
		wg:          wg,
		p:           p,
		finished:    false,
	}
	return progressMgr
}

// adds a new progress bar / spinner with the given name and total work units,
// and starts a goroutine to execute the provided work function.
func (m *ProgressBarMgr) AddBarOrSpinner(barCfg *ProgressBarCfg, work func(interfaces.ProgressIncrement) error) *ProgressBarMgr {
	m.workersChan <- struct{}{} // acquire a worker slot

	// lock to ensure thread safety when adding bars,
	// and to prevent adding bars after Wait() has been called
	m.mu.Lock()
	defer m.mu.Unlock()

	// if the progress manager has already been waited on, it means it's done
	// and should not accept new bars
	if m.finished {
		panic("cannot add bar after Wait() has been called")
	}

	// create the progress bar or spinner and add to WaitGroup
	bar := m.p.New(
		barCfg.total,
		barCfg.style,
		barCfg.options...,
	)
	m.wg.Add(1)

	// create an error channel for this bar to report any error from the work function
	errChan := make(chan error, 1) // buffered channel to avoid blocking
	m.errChans = append(m.errChans, errChan)
	go func() {
		// tickerDone is used to signal the ticker goroutine to stop when the work is done,
		// workErr is used to capture any error from the work function, sending it to errChan
		var workErr error
		defer func() {
			// Recover from panic (avoid crashing the whole process)
			if r := recover(); r != nil {
				workErr = fmt.Errorf("panic in work thread: %v", r)
			}

			// set the progress bar to complete (total) if work succeeded,
			// or to current if there was an error
			if workErr != nil {
				bar.Abort(false)
			} else {
				bar.SetCurrent(barCfg.total)
				bar.SetTotal(barCfg.total, true)
			}

			// report any error from the work function or panic to the error channel
			errChan <- workErr

			// release the worker slot
			<-m.workersChan

			// mark this bar as done in the WaitGroup
			m.wg.Done()
		}()

		// start variable is solely for EWMA calculation
		start := time.Now()

		// the callback will be called by the work function,
		//    which updates the progress bar accordingly
		workErr = work(func(delta int64) {
			// set the current progress
			if !barCfg.IsIndefinite() {
				bar.EwmaIncrBy(int(delta), time.Since(start))
				start = time.Now() // only reset when used
			}
		})
	}()
	return m
}

// Wait waits for all bars to complete and flushes the output.
//
//	reports any error from the work functions of the bars, returning the
//	first error encountered.
func (m *ProgressBarMgr) Wait() error {
	// lock to prevent adding new bars while waiting
	m.mu.Lock()
	defer m.mu.Unlock()

	m.finished = true // mark as finished to prevent adding new bars
	m.p.Wait()        // wait for all bars to complete

	// collect errors from all bars, if any, and return a combined error
	var errGrp error
	for _, errChan := range m.errChans {
		// read the error from the channel
		if err := <-errChan; err != nil {
			errGrp = errors.Join(err, errGrp)
		}

		// close the error channel to prevent goroutine leaks
		close(errChan)
	}
	m.errChans = nil // clear errChans to release references and prevent memory leaks

	return errGrp
}

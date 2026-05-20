// internal/utils/progress_bar.go

package utils

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type ProgressIncrement func(int64)

// ----------------------
// PROGRESS BAR STYLE
// ----------------------

type ProgressBarCfg struct {
	total        int64
	style        mpb.BarFillerBuilder
	prependDecor []decor.Decorator
	appendDecor  []decor.Decorator
}

func NewBarCfg(name string, total int64) ProgressBarCfg {
	nameShort := name[:min(20, len(name))] // truncate name if too long
	decorName := decor.Name(nameShort, decor.WC{
		W: len(nameShort),
		C: decor.DSyncWidthR,
	})
	if total <= 0 {
		return ProgressBarCfg{
			total: -1, // -1 indicates spinner mode
			style: defaultSpinnerStyle(),
			prependDecor: []decor.Decorator{
				decorName,
			},
			appendDecor: []decor.Decorator{
				decor.OnComplete(
					decor.OnAbort(decor.Name(""), " failed"), "   done",
				),
				decor.Name(" | ET: "),
				decor.Elapsed(decor.ET_STYLE_MMSS, decor.WCSyncSpaceR),
			},
		}
	}
	return ProgressBarCfg{
		total: total,
		style: defaultBarStyle(),
		prependDecor: []decor.Decorator{
			decorName,
		},
		appendDecor: []decor.Decorator{
			decor.OnComplete(
				decor.OnAbort(decor.Percentage(decor.WC{W: 5}), " failed"), "   done",
			),
			decor.Name(" | ET: "),
			decor.Elapsed(decor.ET_STYLE_MMSS, decor.WCSyncSpaceR),
			decor.OnCompleteOrOnAbort(decor.Name(" | ETA: "), ""),
			decor.OnCompleteOrOnAbort(decor.EwmaETA(decor.ET_STYLE_MMSS, 30, decor.WCSyncSpaceR), ""),
		},
	}
}

func (c *ProgressBarCfg) IsSpinner() bool {
	return c.total <= 0
}

// defaultBarStyle returns a default style for progress bars, which can be used if no custom
// style is provided.
func defaultBarStyle() mpb.BarStyleComposer {
	return mpb.BarStyle().Lbound("[").Filler("█").Tip("█").Padding("░").Rbound("]")
}

// defaultSpinnerStyle returns a default style for spinner bars.
// The spinner will cycle through the specified characters to indicate progress.
func defaultSpinnerStyle() mpb.SpinnerStyleComposer {
	return mpb.SpinnerStyle(
		">==============<",
		"=>============<=",
		"==>==========<==",
		"===>========<===",
		"====>======<====",
		"=====>====<=====",
		"======>==<======",
		"=======><=======",
		"=======<>=======",
		"======<==>======",
		"=====<====>=====",
		"====<======>====",
		"===<========>===",
		"==<==========>==",
		"=<============>=",
		"<==============>",
	)
}

// ----------------------
// PROGRESS BAR MANAGER
// ----------------------

type ProgressBarMgr struct {
	mu       sync.Mutex
	wg       *sync.WaitGroup
	p        *mpb.Progress
	errChans []chan error
	finished bool
}

// NewProgressBarMgr creates a new ProgressBarMgr with the given style.
func NewProgressBarMgr() *ProgressBarMgr {
	var wg = &sync.WaitGroup{}
	var p = mpb.New(mpb.WithWaitGroup(wg))

	return &ProgressBarMgr{
		wg:       wg,
		p:        p,
		finished: false,
	}
}

// adds a new progress bar / spinner with the given name and total work units,
// and starts a goroutine to execute the provided work function.
func (m *ProgressBarMgr) AddBarOrSpinner(barCfg ProgressBarCfg, work func(ProgressIncrement) error) *ProgressBarMgr {
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
		mpb.PrependDecorators(barCfg.prependDecor...),
		mpb.AppendDecorators(barCfg.appendDecor...),
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

			// mark this bar as done in the WaitGroup
			m.wg.Done()
		}()

		// start variable is solely for EWMA calculation
		start := time.Now()

		// the callback will be called by the work function,
		//    which updates the progress bar accordingly
		workErr = work(func(delta int64) {
			// set the current progress
			if !barCfg.IsSpinner() {
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

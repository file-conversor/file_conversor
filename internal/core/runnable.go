// internal/core/runnable.go

package core

import (
	"fmt"

	"github.com/file-conversor/file_conversor/internal/interfaces"
)

type Runnable struct {
	// Run executes a runnable function, run progress callback (0-100) and returns an error if any
	Func func(interfaces.ProgressIncrement) error

	// list of cleanup functions to run after Func is done
	CleanupFuncs []func() error

	// Output file path (optional, used for progress bar labeling)
	OutputPath string

	// Err is set if there was an error creating the runnable (e.g. invalid input), otherwise nil
	Err error
}

func NewRunnable() *Runnable {
	return &Runnable{
		Func: func(interfaces.ProgressIncrement) error { return nil },
	}
}

func (r *Runnable) SetRun(run func(interfaces.ProgressIncrement) error) {
	r.Func = run
}

func (r *Runnable) SetError(err error) {
	r.Err = err
}

func (r *Runnable) SetOutputPath(path string) {
	r.OutputPath = path
}

func (r *Runnable) AppendCleanup(cleanup ...func() error) {
	r.CleanupFuncs = append(r.CleanupFuncs, cleanup...)
}

func (r *Runnable) Run(updateProgress interfaces.ProgressIncrement) error {
	defer func() {
		if r.CleanupFuncs == nil {
			return // no cleanup funcs to run
		}
		for _, cleanup := range r.CleanupFuncs {
			cleanup() // run any cleanup functions after the main function is done
		}
		r.CleanupFuncs = []func() error{} // clear cleanup funcs after running
	}()
	if r.Err != nil {
		return r.Err // return error if there was an error creating the runnable
	}
	if r.Func == nil {
		return fmt.Errorf("runnable - no function to run") // return error if no function to run
	}
	return r.Func(updateProgress)
}

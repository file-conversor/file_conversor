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

	// Err is set if there was an error creating the runnable (e.g. invalid input), otherwise nil
	Err error
}

func NewRunnable() *Runnable {
	return &Runnable{
		Func: func(interfaces.ProgressIncrement) error { return nil },
	}
}

func (this *Runnable) SetRun(run func(interfaces.ProgressIncrement) error) {
	this.Func = run
}

func (this *Runnable) SetError(err error) {
	this.Err = err
}

func (this *Runnable) AppendCleanup(cleanup ...func() error) {
	this.CleanupFuncs = append(this.CleanupFuncs, cleanup...)
}

func (this *Runnable) Run(updateProgress interfaces.ProgressIncrement) error {
	defer func() {
		if this.CleanupFuncs == nil {
			return // no cleanup funcs to run
		}
		for _, cleanup := range this.CleanupFuncs {
			cleanup() // run any cleanup functions after the main function is done
		}
	}()
	if this.Err != nil {
		return this.Err // return error if there was an error creating the runnable
	}
	if this.Func == nil {
		return fmt.Errorf("runnable - no function to run") // return error if no function to run
	}
	return this.Func(updateProgress)
}

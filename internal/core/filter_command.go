// internal/core/filter_command.go

package core

import (
	"errors"
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/interfaces"
)

type FilterCommand struct {
	flags.InputFilesArg // input file arguments (paths, recurse flag, batch file)
}

func (m *FilterCommand) Parse(formats interfaces.FormatInterface) error {
	if err := errors.Join(
		m.InputFilesArg.Parse(formats),
	); err != nil {
		return fmt.Errorf("parse files: %w", err)
	}
	return nil
}

func (m *FilterCommand) Validate(formats interfaces.FormatInterface) error {
	if err := errors.Join(
		m.InputFilesArg.Validate(formats),
	); err != nil {
		return fmt.Errorf("validate files: %w", err)
	}
	return nil
}

// Processes the input files and creates a Runnable for each file to be processed.
func (m *FilterCommand) GetRunnable(
	run func(
		inFile *os.File,
		runnable *Runnable,
		updateProgress interfaces.ProgressIncrement,
	) error,
) <-chan *Runnable {
	outChan := make(chan *Runnable) // channel to send runnable to be executed

	go func() {
		defer close(outChan) // close channel when done

		processFileFunc := func(inFile *os.File, cleanup func() error) error {
			runnable := NewRunnable()
			runnable.AppendCleanup(cleanup) // ensure cleanup is called after processing

			// use input file path as output path (for check validation messages)
			runnable.SetOutputPath(inFile.Name())

			// call the specific command's run method
			runFunc := func(updateProgress interfaces.ProgressIncrement) error {
				return run(inFile, runnable, updateProgress)
			}
			runnable.SetRun(runFunc)

			// send runnable to be executed
			outChan <- runnable
			return nil
		}

		// process input files
		if err := env.OpenInputFiles(processFileFunc, m.InputFiles...); err != nil {
			// if there's an error opening input files, send a runnable with the error
			runnable := NewRunnable()
			runnable.SetError(err)

			// send runnable with error to be executed
			outChan <- runnable
		}
	}()

	return outChan
}

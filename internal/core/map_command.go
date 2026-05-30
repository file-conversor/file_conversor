// internal/core/map_command.go

package core

import (
	"errors"
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type MapCommand struct {
	flags.OutputDirFlag // output dir flag (dir, overwrite, suffix, format)
	flags.InputFilesArg // input file arguments (paths, recurse flag, batch file)
}

func (m *MapCommand) Parse(formats interfaces.FormatInterface) error {
	if err := errors.Join(
		m.OutputDirFlag.Parse(formats),
		m.InputFilesArg.Parse(formats),
	); err != nil {
		return fmt.Errorf("parse files: %w", err)
	}
	return nil
}

func (m *MapCommand) Validate(formats interfaces.FormatInterface) error {
	if err := errors.Join(
		m.OutputDirFlag.Validate(formats),
		m.InputFilesArg.Validate(formats),
	); err != nil {
		return fmt.Errorf("validate files: %w", err)
	}
	return nil
}

// Processes the input files and creates a Runnable for each file to be processed.
func (m *MapCommand) GetRunnable(
	run func(
		inFile *os.File,
		outFile *os.File,
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

			// open output file
			outFile, err := m.GetOutputFile(inFile.Name())
			if err != nil {
				return fmt.Errorf("MapCommand - open output file for '%s': %w", inFile.Name(), err)
			}
			runnable.AppendCleanup(outFile.Close) // ensure output file is closed after processing
			runnable.SetOutputPath(outFile.Name())

			// call the specific command's run method
			runFunc := func(updateProgress interfaces.ProgressIncrement) error {
				err := run(inFile, outFile, runnable, updateProgress)
				if err != nil {
					// if there's an error during processing, ensure the output file is removed
					runnable.AppendCleanup(func() error {
						logger.Warnf("Removing output file '%s'\n", outFile.Name())
						return os.Remove(outFile.Name())
					})
				}
				return err
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

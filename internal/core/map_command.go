// internal/core/map_command.go

package core

import (
	"errors"
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/core/flags"
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
		inFile string,
		outFile string,
		runnable *Runnable,
		updateProgress interfaces.ProgressIncrement,
	) error,
) <-chan *Runnable {
	outChan := make(chan *Runnable) // channel to send runnable to be executed

	go func() {
		defer close(outChan) // close channel when done

		for _, inFile := range m.InputFiles {
			runnable := NewRunnable()
			outFile, err := m.GetOutputPath(inFile)
			if err != nil {
				runnable.SetError(fmt.Errorf("MapCommand - get output filename for '%s': %w", inFile, err))
				outChan <- runnable
				return
			}
			runnable.SetOutputPath(outFile)

			// call the specific command's run method
			runFunc := func(updateProgress interfaces.ProgressIncrement) error {
				err := run(inFile, outFile, runnable, updateProgress)
				if err != nil {
					// if there's an error during processing, ensure the output file is removed
					runnable.AppendCleanup(func() error {
						logger.Warnf("Removing output file '%s'\n", outFile)
						return os.Remove(outFile)
					})
				}
				return err
			}
			runnable.SetRun(runFunc)

			// send runnable to be executed
			outChan <- runnable
		}
	}()

	return outChan
}

// internal/core/reduce_command.go

package core

import (
	"errors"
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
	"github.com/file-conversor/file_conversor/internal/validation"
)

type ReduceCommand struct {
	Append               bool // append to output file (if it exists)
	flags.OutputFileFlag      // output file arguments (path, overwrite flag)
	flags.InputFilesArg       // input file arguments (paths, recurse flag, batch file)
}

func (r *ReduceCommand) Parse(formats interfaces.FormatInterface) error {
	// if append is true, we need to allow overwriting the output file
	r.OutputFileFlag.Overwrite = r.OutputFileFlag.Overwrite || r.Append

	if err := errors.Join(
		r.OutputFileFlag.Parse(formats),
		r.InputFilesArg.Parse(formats),
	); err != nil {
		return fmt.Errorf("parse files: %w", err)
	}
	return nil
}

func (r *ReduceCommand) Validate(formats interfaces.FormatInterface) error {
	if err := errors.Join(
		r.OutputFileFlag.Validate(formats),
		r.InputFilesArg.Validate(formats),
		validation.InputOutputNotEqual(r.OutputFile, r.InputFiles...),
	); err != nil {
		return fmt.Errorf("validate files: %w", err)
	}
	return nil
}

// Processes the input files and creates a Runnable for each file to be processed.
func (r *ReduceCommand) GetRunnable(
	run func(
		inFiles []string,
		outFile string,
		runnable *Runnable,
		updateProgress interfaces.ProgressIncrement,
	) error,
) <-chan *Runnable {
	runnable := NewRunnable()
	outChan := make(chan *Runnable) // channel to send runnable to be executed

	go func() {
		defer close(outChan) // close channel when done

		var inFiles []string

		// append existing output file (if exists)
		if r.Append && env.FileExists(r.OutputFile) {
			// copy existing output file to tmp file (if it exists) so that we can append it to output
			tmpFile, cleanup, err := env.CopyToTmpFile(r.OutputFile)
			if err != nil {
				runnable.SetError(fmt.Errorf("copy output to tmp file: %w", err))
				return
			}
			tmpFile.Close()                           // close tmp file handle (we only need its path for appending to output)
			runnable.AppendCleanup(cleanup)           // ensure tmp file is cleaned up
			inFiles = append(inFiles, tmpFile.Name()) // add tmp file as input to be merged with the other input files
		}

		// append input files
		inFiles = append(inFiles, r.InputFiles...)

		// call the specific command's run method
		runFunc := func(updateProgress interfaces.ProgressIncrement) error {
			err := run(inFiles, r.OutputFile, runnable, updateProgress)
			if err != nil {
				// if there's an error during processing, ensure the output file is removed
				runnable.AppendCleanup(func() error {
					logger.Warnf("Removing output file '%s'\n", r.OutputFile)
					return os.Remove(r.OutputFile)
				})
			}
			return err
		}
		runnable.SetRun(runFunc)

		outChan <- runnable // send runnable
	}()
	return outChan
}
